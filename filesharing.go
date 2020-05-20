package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/Mikhalevich/file_service/proto"
	"github.com/Mikhalevich/filesharing/handlers"
	"github.com/Mikhalevich/filesharing/router"
	"github.com/Mikhalevich/goauth"
	"github.com/Mikhalevich/goauth/db"
	"github.com/Mikhalevich/goauth/email"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type params struct {
	Host         string
	DBConnection string
}

func loadParams() (*params, error) {
	var p params

	p.Host = os.Getenv("FS_HOST")
	if p.Host == "" {
		return nil, errors.New("host name is empty, please specify FS_HOST variable")
	}

	p.DBConnection = os.Getenv("FS_DB_CONNECTION_STRING")
	// if p.DBConnection == "" {
	// 	return nil, errors.New("[loadParams] database connection string is empty, please specify FS_DB_CONNECTION_STRING variable")
	// }

	return &p, nil
}

func main() {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	params, err := loadParams()
	if err != nil {
		logger.Error(err)
		return
	}

	storageChecker := router.NewPublicStorages()

	var auth goauth.Authentifier
	enableAuth := params.DBConnection != ""
	if enableAuth {
		var pg *db.Postgres
		for i := 0; i < 3; i++ {
			pg, err = db.NewPostgres(params.DBConnection)
			if err == nil {
				break
			}

			logger.Infof("attemp connect to database: %d  error: %v", i, err)
		}

		if err != nil {
			logger.Error(err)
			return
		}

		defer pg.Close()

		es := &email.GomailSender{
			Host:     "smtp.gmail.com",
			Port:     587,
			From:     "",
			Password: "",
		}
		auth = goauth.NewAuthentificator(pg, pg, NewCookieSession(storageChecker, 1*60*60*24*30), es)
	} else {
		auth = goauth.NewNullAuthentificator()
	}

	fsConnection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		logger.Errorf("did not connect: %w", err)
		return
	}
	defer fsConnection.Close()
	fsClient := proto.NewFileServiceClient(fsConnection)

	h := handlers.NewHandlers(storageChecker, auth, NewGRPCFileServiceClient(fsClient), logger)
	r := router.NewRouter(enableAuth, h, logger)

	logger.Infof("Running params = %s", params)

	err = http.ListenAndServe(params.Host, r.Handler())
	if err != nil {
		logger.Error(err)
	}
}
