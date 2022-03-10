package snippet

// Sample: https://vincent.composieux.fr/article/dependency-injection-in-go-with-uber-go-fx
// Doc: https://pmihaylov.com/shared-components-go-microservices/
// Source: https://github.com/preslavmihaylov/fxappexample/tree/v2-fx-example
// fx module: https://github.com/preslavmihaylov/fxappexample/tree/v3-fx-modules
import (
	"go.uber.org/zap"
	"net/http"
)

func main() {
	fx.New(
		fx.Provide(ProvideConfig),
		fx.Provide(ProvideLogger),
		fx.Provide(http.NewServeMux),
		fx.Invoke(httphandler.New),
		fx.Invoke(registerHooks),
	).Run()
}

func registerHooks(
	lifecycle fx.Lifecycle,
	logger *zap.SugaredLogger, cfg *Config, mux *http.ServeMux,
) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				logger.Info("Listening on ", cfg.ApplicationConfig.Address)
				go http.ListenAndServe(cfg.ApplicationConfig.Address, mux)
				return nil
			},
			OnStop: func(context.Context) error {
				return logger.Sync()
			},
		},
	)
}
