/*
	Copyright (C) 2018 Nirmal Almara

    This file is part of Joyread.

    Joyread is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Joyread is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
	along with Joyread.  If not, see <https://www.gnu.org/licenses/>.
*/

package joyread

import (
	// built-in packages
	"database/sql"
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"strconv" // vendor packages

	"github.com/gin-gonic/gin" // custom packages
	"gitlab.com/joyread/server/books"
	cError "gitlab.com/joyread/server/error"
	"gitlab.com/joyread/server/home"
	"gitlab.com/joyread/server/middleware"
	"gitlab.com/joyread/server/models"
	"gitlab.com/joyread/server/onboard"
	"gitlab.com/joyread/server/settings"
)

// StartServer handles the URL routes and starts the server
func StartServer() {
	// Gin initiate
	r := gin.Default()

	fmt.Println(reflect.TypeOf(r))

	conf := settings.GetConf()

	// Serve static files
	r.Static("/assets", path.Join(conf.BaseValues.AssetPath, "assets"))

	// HTML rendering
	r.LoadHTMLGlob(path.Join(conf.BaseValues.AssetPath, "assets/templates/*"))

	// Open postgres database
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", conf.BaseValues.DBValues.DBUsername, conf.BaseValues.DBValues.DBPassword, conf.BaseValues.DBValues.DBHostname, conf.BaseValues.DBValues.DBPort, conf.BaseValues.DBValues.DBName, conf.BaseValues.DBValues.DBSSLMode)
	db, err := sql.Open("postgres", connStr)
	cError.CheckError(err)
	defer db.Close()

	fmt.Println(runtime.GOOS)
	fmt.Println(runtime.GOARCH)

	if runtime.GOOS == "windows" {
		fmt.Println("Hello from Windows")
	}

	// models.CreateLegend(db)
	models.CreateAccount(db)
	models.CreateBooks(db)
	// models.CreateSMTP(db)
	// models.CreateNextcloud(db)

	r.Use(
		middleware.CORSMiddleware(),
		middleware.APIMiddleware(db),
		middleware.UserMiddleware(db),
	)

	// Gin handlers
	r.GET("/", home.Home)
	r.GET("/uploads/:bookName", home.ServeBook)
	r.GET("/cover/:coverName", home.ServeCover)
	r.GET("/signin", home.Home)
	r.GET("/send-file", home.SendFile)
	r.POST("/signin", onboard.PostSignIn)
	r.GET("/signup", onboard.GetSignUp)
	r.POST("/signup", onboard.PostSignUp)
	r.GET("/signout", onboard.SignOut)
	r.GET("/storage", onboard.GetStorage)
	r.POST("/nextcloud", onboard.PostNextcloud)
	r.GET("/nextcloud-auth/:user_id", onboard.NextcloudAuthCode)
	r.POST("/upload-books", books.UploadBooks)
	r.GET("/book/:bookName", books.GetBook)
	r.GET("/viewer/:bookName", books.Viewer)

	// Listen and serve
	port, err := strconv.Atoi(conf.BaseValues.ServerPort)
	if err != nil {
		fmt.Println("Invalid port specified")
		os.Exit(1)
	}
	r.Run(fmt.Sprintf(":%d", port))
}