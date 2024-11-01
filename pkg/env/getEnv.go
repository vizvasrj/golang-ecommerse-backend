package env

// type Env struct {
// 	DBName             string
// 	DBUri              string
// 	SecretJWT          string
// 	GoogleClientID     string
// 	GoogleClientSecret string
// 	GoogleRedirectURL  string
// 	ClientURL          string
// }

// func GetEnv() (*Env, error) {
// 	db_name := os.Getenv("DB_NAME")
// 	if db_name == "" {
// 		return nil, errors.New("DB_NAME is not set")
// 	}
// 	db_uri := os.Getenv("DB_URI")
// 	if db_uri == "" {
// 		return nil, errors.New("env not found DB_URI")
// 	}

// 	SecretJWT := os.Getenv("SECRET_JWT")
// 	if SecretJWT == "" {
// 		return nil, errors.New("env not found SECRET_JWT")
// 	}
// 	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
// 	if googleClientID == "" {
// 		return nil, errors.New("env not found GOOGLE_CLIENT_ID")
// 	}
// 	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
// 	if googleClientSecret == "" {
// 		return nil, errors.New("env not found GOOGLE_CLIENT_SECRET")
// 	}
// 	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
// 	if googleRedirectURL == "" {
// 		return nil, errors.New("env not found GOOGLE_REDIRECT_URL")
// 	}

// 	clientURL := os.Getenv("CLIENT_URL")
// 	if clientURL == "" {
// 		return nil, errors.New("env not found CLIENT_URL")
// 	}

// 	return &Env{
// 		DBName:             db_name,
// 		DBUri:              db_uri,
// 		SecretJWT:          SecretJWT,
// 		GoogleClientID:     googleClientID,
// 		GoogleClientSecret: googleClientSecret,
// 		GoogleRedirectURL:  googleRedirectURL,
// 		ClientURL:          clientURL,
// 	}, nil

// }

import (
	"github.com/kelseyhightower/envconfig"
)

type Env struct {
	DBName             string `envconfig:"DB_NAME" required:"true"`
	DBUri              string `envconfig:"DB_URI" required:"true"`
	SecretJWT          string `envconfig:"SECRET_JWT" required:"true"`
	GoogleClientID     string `envconfig:"GOOGLE_CLIENT_ID" required:"true"`
	GoogleClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET" required:"true"`
	GoogleRedirectURL  string `envconfig:"GOOGLE_REDIRECT_URL" required:"true"`
	ClientURL          string `envconfig:"CLIENT_URL" required:"true"`
	AWSAccessKeyID     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	AWSSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
	AWSRegion          string `envconfig:"AWS_REGION" required:"true"`
	AWSBucketName      string `envconfig:"AWS_BUCKET_NAME" required:"true"`
	AWSEndpoint        string `envconfig:"AWS_ENDPOINT" required:"true"`
}

func GetEnv() (*Env, error) {
	var env Env
	err := envconfig.Process("", &env)
	if err != nil {
		return nil, err
	}
	return &env, nil
}
