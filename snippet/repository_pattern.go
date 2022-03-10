package snippet

package repositories

import "go.uber.org/zap"

type GenericDataRepo interface {
	CloneWithTransID(transID int) GenericDataRepo
	ListUsers() []string
}

type DataRepo struct {
	logger *zap.Logger
	db     *FakeDB
}

func NewDataRepo() DataRepo {
	db := connectToDb()
	logger, _ := zap.NewProduction()
	return DataRepo{logger: logger, db: &db}
}

func (d DataRepo) CloneWithTransID(transID int) GenericDataRepo {
	newLogger := d.logger.With(zap.Int("transID", transID))
	return DataRepo{logger: newLogger, db: d.db}
}

func (d DataRepo) ListUsers() []string {
	d.logger.Info("getting users")
	return d.db.GetUsers()
}

type FakeDB struct{}

func (m FakeDB) GetUsers() []string {
	return []string{"amy", "bob", "carl"}
}

func connectToDb() FakeDB {
	return FakeDB{}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.TransactionIdMiddleware)
	dataRepo := repositories.NewDataRepo()

	handler := handlers.NewHelloWorldHandler(dataRepo)

	r.Method(GET, "/", handler)
	http.ListenAndServe(":3000", r)
}