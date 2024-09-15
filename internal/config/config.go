package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct{
	Env string `yaml:"env" env-required:"true"`
	TokenTTL time.Duration `yaml:"token_ttl" env-required:"true"`
	GRPC GRPCConfig `yaml:"grpc" env-required:"true"`
	DB DB
}

type DB struct{
	User string
	Password  string
	Host string
	Port string
	Name string
}

type GRPCConfig struct{
	Port int `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}


func MustLoad() *Config{
	//  configPath := fetchConfigPath()
	configPath := "/home/satori/GO/src-/src/grpc/sso/config/local.yaml"

	if _, err:=os.Stat(configPath); os.IsNotExist(err){
		panic("config file doesn't exist " + configPath)
	}

	var cfg Config
	if err:=cleanenv.ReadConfig(configPath, &cfg); err!=nil{
		panic("failed to read config: " + err.Error())
	}

	// mustReadEnv(&cfg)
	cfg.DB.Name = "grpc_auth"
	cfg.DB.User = "grpc"
	cfg.DB.Password = "grpc"
	cfg.DB.Port="5432"
	cfg.DB.Host = "localhost"

	return &cfg
}

// func mustReadEnv(cfg *Config){
// 	cfg.DB.User = os.Getenv("DB_USERNAME")
// 	if cfg.DB.User==""{
// 		panic(`can't find "DB_USERNAME" env`)
// 	}
// 	cfg.DB.User = os.Getenv("DB_PASSWORD")
// 	if cfg.DB.User==""{
// 		panic(`can't find "DB_PASSWORD" env`)
// 	}	
// 	cfg.DB.User = os.Getenv("DB_HOST")
// 	if cfg.DB.User==""{
// 		panic(`can't find "DB_HOST" env`)
// 	}	
// 	cfg.DB.User = os.Getenv("DB_PORT")
// 	if cfg.DB.User==""{
// 		panic(`can't find "DB_PORT" env`)
// 	}	
// 	cfg.DB.User = os.Getenv("DB_NAME")
// 	if cfg.DB.User==""{
// 		panic(`can't find "DB_NAME" env`)
// 	}

// }

func fetchConfigPath() string{
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path==""{
		path = os.Getenv("CONFIG_PATH")
	}

	if path==""{
		panic(`Can't find config, try flag --config="path/to/config.yaml" or use "CONFIG_PATH" environment`)
	}

	return path
}