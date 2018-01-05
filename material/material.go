package material

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/qor/auth"
	"github.com/qor/auth/claims"
	"github.com/qor/auth/providers/password"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/render"
	"github.com/qor/l10n"
)

// ErrPasswordConfirmationNotMatch password confirmation not match error
var ErrPasswordConfirmationNotMatch = errors.New("password confirmation doesn't match password")

func currentLocale(req *http.Request) string {
	locale := l10n.Global
	if cookie, err := req.Cookie("locale"); err == nil {
		locale = cookie.Value
	}
	return locale
}

// New initialize clean theme
func New(config *auth.Config) *auth.Auth {
	if config == nil {
		config = &auth.Config{}
	}
	config.ViewPaths = append(config.ViewPaths, "github.com/grengojbo/auth_themes/material/views")

	if config.DB == nil {
		fmt.Print("Please configure *gorm.DB for Auth theme clean")
	}

	if config.Render == nil {
		yamlBackend := yaml.New()
		I18n := i18n.New(yamlBackend)
		locales := []string{"ru-RU", "uk-UA", "ru-RU", "pl", "en-US"}
		for _, gopath := range append([]string{filepath.Join(utils.AppRoot, "vendor")}, utils.GOPATH()...) {
			for _, localeName := range locales {
				fileLocale := fmt.Sprintf("github.com/grengojbo/auth_themes/material/locales/%s.yml", localeName)
				filePath := filepath.Join(gopath, "src", fileLocale)
				if content, err := ioutil.ReadFile(filePath); err == nil {
					//fmt.Println("Read locale:", localeName)
					translations, _ := yamlBackend.LoadYAMLContent(content)
					for _, translation := range translations {
						I18n.AddTranslation(translation)
					}
					//break
				}
			}
		}

		config.Render = render.New(&render.Config{
			//ViewPaths:     []string{"app/views"},
			//DefaultLayout: "layouts/application",
			FuncMapMaker: func(render *render.Render, req *http.Request, w http.ResponseWriter) template.FuncMap {
				return template.FuncMap{
					"t": func(key string, args ...interface{}) template.HTML {
						//fmt.Println("--> Locale:", utils.GetLocale(&qor.Context{Request: req}))
						return I18n.T(utils.GetLocale(&qor.Context{Request: req}), key, args...)
					},
					"current_locale": func() string {
						return currentLocale(req)
					},
				}
			},
		})
	}

	Auth := auth.New(config)

	Auth.RegisterProvider(password.New(&password.Config{
		Confirmable: true,
		RegisterHandler: func(context *auth.Context) (*claims.Claims, error) {
			context.Request.ParseForm()

			if context.Request.Form.Get("confirm_password") != context.Request.Form.Get("password") {
				return nil, ErrPasswordConfirmationNotMatch
			}

			return password.DefaultRegisterHandler(context)
		},
	}))

	if Auth.Config.DB != nil {
		// Migrate Auth Identity model
		Auth.Config.DB.AutoMigrate(Auth.Config.AuthIdentityModel)
	}
	return Auth
}
