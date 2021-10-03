package main

import (
	"context"
	redisv8 "github.com/go-redis/redis/v8"
	xhttp "github.com/goclub/http"
	red "github.com/goclub/redis"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

func main () {
	VERSION := "v0.0.1"
	log.Print(VERSION)
	ctx := context.Background()
	configBytes, err := ioutil.ReadFile(path.Join(os.Getenv("GOPATH"), "src/work/echo/env", "config.yaml"))
	if err != nil {
		log.Print(err) //为了演示故意不panic
	}
	config := struct {
		RedisAddr string `yaml:"REDIS_ADDR"`
	}{}
	err = yaml.Unmarshal(configBytes, &config) ; if err != nil {
		log.Print(err) //为了演示故意不panic
	}
	log.Printf("config %+v", config)
	redisCoreClient := redisv8.NewClient(&redisv8.Options{
		Network: "tcp",
		Addr: config.RedisAddr,
	})
	redisClient := red.GoRedisV8{
		Core: redisCoreClient,
	}
	err = redisClient.Core.Ping(ctx).Err() ; if err != nil {
		log.Print(err) //为了演示故意不panic
	}
	r := xhttp.NewRouter(xhttp.RouterOption{})
	r.Use(func(c *xhttp.Context, next xhttp.Next) (reject error) {
		requestTime := time.Now()
		log.Print("Request: ", c.Request.Method, c.Request.URL.String())
		reject = next() ; if reject != nil {
			return
		}
		responseTime := time.Now().Sub(requestTime)
		log.Print("Response: (" , responseTime.String(), ") ", c.Request.Method, c.Request.URL.String())
		return nil
	})
	r.HandleFunc(xhttp.Route{xhttp.GET, "/"}, func(c *xhttp.Context) (err error) {
		query := c.Request.URL.Query()
		reply := map[string]string{}
		for key, value := range query {
			reply[key] = value[0]
		}
		return c.WriteJSON(reply)
	})
	r.HandleFunc(xhttp.Route{xhttp.GET, "/version"}, func(c *xhttp.Context) (err error) {
		return c.WriteJSON(VERSION)
	})
	// 用于测试pod退出自动重启
	r.HandleFunc(xhttp.Route{xhttp.GET, "/debug/exit"}, func(c *xhttp.Context) (err error) {
		os.Exit(0)
		return nil
	})
	// 用于测试连接k8s中其他service
	r.HandleFunc(xhttp.Route{xhttp.GET, "/count"}, func(c *xhttp.Context) (err error) {
		ctx := c.Request.Context()
		query := c.Request.URL.Query()
		key := query.Get("key")
		if key == "" {
			key = "default"
		}
		newValue, err := red.INCR{
			Key: key,
		}.Do(ctx, redisClient) ; if err != nil {
		    return
		}
		err = c.WriteJSON(map[string]interface{}{
			"newValue": newValue,
		}) ; if err != nil {
		    return
		}
		return
	})
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}
	addr := ":" + serverPort
	server := &http.Server{
		Addr: addr,
		Handler: r,
	}
	r.LogPatterns(server)
	go func() {
		log.Print(server.ListenAndServe())
	}()
	xhttp.GracefulClose(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
		defer cancel()
		log.Print(redisCoreClient.Close())
		log.Print(server.Shutdown(ctx))
	})
}
