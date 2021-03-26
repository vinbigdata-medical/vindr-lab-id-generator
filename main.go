package main

import (
	"fmt"
	"id-generator/constants"
	"id-generator/generator"
	"id-generator/utils"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rqlite/gorqlite"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	_ "github.com/rqlite/gorqlite"
)

func newLogger() *zap.Logger {
	env := viper.GetString("workspace.env")
	var logger *zap.Logger
	switch env {
	case "DEVELOPMENT":
		logger, _ = zap.NewDevelopment()
	default:
		logger, _ = zap.NewProduction()
	}
	return logger
}

func initConfigs(env string) {
	viper.AddConfigPath("conf")
	viper.SetConfigName(fmt.Sprintf("config.%s", env))
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "__")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil {
		utils.LogInfo("Error reading config file, %s", err)
	}
	// viper.SetDefault("rqlite.uri", "http://127.0.0.1:4001")
}

func getMapEnvVars() *map[string]string {
	ret := make(map[string]string)
	envsOS := os.Environ()
	for _, envOS := range envsOS {
		items := strings.Split(envOS, "=")
		if len(items) > 1 {
			ret[items[0]] = items[1]
		}
	}
	return &ret
}

func main() {

	envVars := getMapEnvVars()
	env := "development"
	if value, found := (*envVars)[constants.ApiOsEnv]; found {
		env = value
	}
	utils.LogInfo(fmt.Sprintf("API is running in [%s] mode", env))
	initConfigs(env)
	fmt.Println(viper.GetString("rqlite.uri"))

	route := gin.Default()
	logger := newLogger()

	route.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "PUT", "GET", "DELETE"},
		AllowHeaders:     []string{"Access-Control-Allow-Headers", "Origin", "Accept", "X-Requested-With", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	var conn gorqlite.Connection
	rqliteURI := viper.GetString("rqlite.uri")

	for true {
		conn1, err := gorqlite.Open(rqliteURI)

		_, err1 := conn1.Leader()
		utils.LogInfo("%s\t%s\t%v", "leader", rqliteURI, err1)

		if err != nil {
			utils.LogInfo("RETRY")
			time.Sleep(1 * time.Second)
		} else {
			utils.LogInfo("CONNECTED")
			conn = conn1
			break
		}
	}

	generatorStore := generator.NewGeneratorStore(&conn, logger)
	generatorAPI := generator.NewGeneratorAPI(generatorStore, logger)
	generatorAPI.InitRoute(route, "/id_generator")

	//cleaner := workers.NewCleaner(generatorStore, logger)
	//go cleaner.cleaner()
	route.Run("0.0.0.0:8080")
}
