package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/jinzhu/now"
)

// Wish 许愿词期待的JSON
type Wish struct {
	CheerWord string `form:"cheer_word" json:"cheer_word"`
	Address   string `form:"address" json:"address" binding:"required"`
	Code      string `form:"code" json:"code" binding:"required"`
	LovePower string `form:"love_power" json:"love_power" binding:"required"`
}

// WishResult 许愿结果
type WishResult struct {
	Success bool   `json:"success"`
	Hard    int64  `json:"hard"`
	Type    string `json:"type"`
	Amount  int64  `json:"amount"`
	Stock   string `json:"stock"`
}

// HandleCriticalError 处理致命错误
func HandleCriticalError(err error) {
	fmt.Println("occurred error:", err)
	os.Exit(-1)
}

// HandleError 处理普通错误
func HandleError(err error) {
	fmt.Println("occurred error:", err)
}

// Timestamp 从当前分钟的第一秒生成需要的时间戳，
func Timestamp() int64 {
	return now.BeginningOfMinute().Unix()
}

// RawOre 根据条件拼接字符串用来产生哈希值
func RawOre(wish Wish) []byte {
	ore := bytes.Join([][]byte{
		[]byte(wish.CheerWord),
		[]byte(wish.Address),
		[]byte(wish.LovePower),
		[]byte(strconv.FormatInt(Timestamp(), 10)),
		[]byte(wish.Code),
	}, []byte{})
	return ore
}

// Hash 生成哈希值
func Hash(ore []byte) [64]byte {
	return sha512.Sum512(ore)
}

// MatchWish 匹配是否满足标准
func MatchWish(hard int, ore []byte) bool {
	bin := Hash(ore)
	zero := (hard / 8)
	for index := 1; index <= zero; index++ {
		if bin[len(bin)-index] != 0 {
			return false
		}
	}

	residual := (hard % 8)

	if residual > 0 {
		last := bin[len(bin)-(hard/8)-1]
		head := fmt.Sprintf("%08b", last)

		if len(head) < residual {
			return false
		}

		headZero := ""
		for index := 0; index < residual; index++ {
			headZero += "0"
		}

		return head[len(head)-residual:] == headZero
	}
	return true
}

// PostWish 发送成功的挖矿请求
func PostWish(wish Wish) WishResult {
	formData := url.Values{
		"cheer_word": {wish.CheerWord},
		"address":    {wish.Address},
		"code":       {wish.Code},
		"love_power": {wish.LovePower},
	}
	resp, err := http.PostForm(os.Getenv("WISH_URL"), formData)
	if err != nil {
		HandleError(err)
		return WishResult{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var res WishResult
	if err := json.Unmarshal(body, &res); err == nil {
		return res
	}
	HandleError(err)
	return res
}

// RedisClient 获取Redis客户端
func RedisClient() *redis.Client {
	db, _ := strconv.ParseInt(os.Getenv("REDIS_DB"), 10, 64)
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PW"),
		DB:       int(db),
	})

	_, err := client.Ping().Result()
	if err != nil {
		HandleCriticalError(err)
	}
	return client
}

// CurrentHard 当前难度
func CurrentHard(client *redis.Client) int64 {
	val, err := client.Get("global:hard").Result()
	if err != nil {
		return 26
	}
	hard, _ := strconv.ParseInt(val, 10, 64)
	return hard
}

func main() {
	client := RedisClient()

	r := gin.Default()

	r.GET("/api/super_wishs", func(c *gin.Context) {
		hard := CurrentHard(client)
		c.JSON(http.StatusOK, gin.H{
			"hard":      hard,
			"unix_time": Timestamp(),
		})
	})

	r.POST("/api/super_wishs", func(c *gin.Context) {
		var form Wish
		hard := int(CurrentHard(client))
		if err := c.ShouldBind(&form); err == nil {
			ore := RawOre(form)
			if MatchWish(hard, ore) {
				res := PostWish(form)
				c.JSON(http.StatusOK, gin.H{
					"success": res.Success,
					"hard":    res.Hard,
					"type":    res.Type,
					"amount":  res.Amount,
					"stock":   res.Stock,
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"msg":     "Sorry, try more time!",
					"hard":    hard,
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

	})

	r.Run(":8000") // listen and serve on 0.0.0.0:8080
}
