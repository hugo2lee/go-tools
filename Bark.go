package main

import (
	"net/http"
	"fmt"
	"log"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"

	"github.com/boltdb/bolt"

	"errors"
	"encoding/json"
	"github.com/go-zoo/bone"
	"github.com/renstrom/shortuuid"
	"strconv"
	"flag"
	"strings"
)

type BaseResponse struct {
	Code    int         `json:"code"`
	Data interface{} `json:"data"`
	Message string      `json:"message"`
}

func responseString(code int, message string)string {
	t, _ := json.Marshal(BaseResponse{Code:code,Message:message})
	return string(t)
}
func responseData(code int, data map[string]interface{} , message string)string {
	t, _ := json.Marshal(BaseResponse{Code:code,Data:data, Message:message})
	return string(t)
}

func ping(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fmt.Fprint(w, responseData(200, map[string]interface{}{"version": "1.0.0"},"pong"))
}

func Index(w http.ResponseWriter, r *http.Request) {
	key := bone.GetValue(r, "key")
	category := bone.GetValue(r, "category")
	title := bone.GetValue(r, "title")
	body := bone.GetValue(r, "body")

	defer r.Body.Close()

	deviceToken ,err := getDeviceTokenByKey(key)
	if err != nil {
		log.Println("找不到key对应的DeviceToken key: " + key)
		fmt.Fprint(w, responseString(400,"找不到key对应的DeviceToken, 请确保Key正确! Key可在App端注册获得。"))
		return
	}

	r.ParseForm()

	if len(title) <= 0 && len(body) <= 0 {
		//url中不包含 title body，则从Form里取
		for key,value := range r.Form{
			if strings.ToLower(key) == "title" {
				title = value[0]
			} else if strings.ToLower(key) == "body"{
				body = value[0]
			}
		}


	}

	if len(body) <= 0 {
		body = "无推送文字内容"
	}

	params := make(map[string]interface{})
	for key,value := range r.Form {
		params[strings.ToLower(key)] = value[0]
	}

	log.Println(" ========================== ")
	log.Println("key: ", key)
	log.Println("category: ", category)
	log.Println("title: ", title)
	log.Println("body: ", body)
	log.Println("params: ", params)
	log.Println(" ========================== ")

	err = postPush(category,title,body,deviceToken,params)
	if err != nil {
		fmt.Fprint(w, responseString(400, err.Error()))
	} else{
		fmt.Fprint(w, responseString(200, ""))
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.ParseForm()
	key := shortuuid.New()
	var deviceToken string
	for key,value := range r.Form {
		if strings.ToLower(key) == "devicetoken" {
			deviceToken = value[0]
			break
		}
	}

	if len(deviceToken) <= 0 {
		fmt.Fprint(w, responseString(400, "deviceToken 不能为空"))
		return
	}

	oldKey := r.FormValue("key")
	boltDB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("device"))
		if err != nil {
			return  err
		}

		if len(oldKey) >0 {
			//如果已经注册，则更新DeviceToken的值
			val := bucket.Get([]byte(oldKey))
			if val != nil {
				key = oldKey
			}
		}

		bucket.Put([]byte(key), []byte(deviceToken))
		return nil
	})
	log.Println("注册设备成功")
	log.Println("key: ", key)
	log.Println("deviceToken: ", deviceToken)
	fmt.Fprint(w, responseData(200, map[string]interface{}{"key":key}, "注册成功"))
}


func getDeviceTokenByKey(key string) (string,error){
	var deviceTokenBytes []byte
	err := boltDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("device"))
		deviceTokenBytes = bucket.Get([]byte(key))
		if deviceTokenBytes == nil {
			return errors.New("没找到DeviceToken")
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return string(deviceTokenBytes), nil
}

func getb() []byte {
	//测试证书
	if IsDev{
		return []byte{48, 130, 12, 81, 2, 1, 3, 48, 130, 12, 24, 6, 9, 42, 134, 72, 134, 247, 13, 1, 7, 1, 160, 130, 12, 9, 4, 130, 12, 5, 48, 130, 12, 1, 48, 130, 6, 151, 6, 9, 42, 134, 72, 134, 247, 13, 1, 7, 6, 160, 130, 6, 136, 48, 130, 6, 132, 2, 1, 0, 48, 130, 6, 125, 6, 9, 42, 134, 72, 134, 247, 13, 1, 7, 1, 48, 28, 6, 10, 42, 134, 72, 134, 247, 13, 1, 12, 1, 6, 48, 14, 4, 8, 47, 8, 74, 91, 186, 62, 186, 103, 2, 2, 8, 0, 128, 130, 6, 80, 53, 244, 199, 61, 44, 252, 47, 243, 72, 173, 63, 42, 8, 90, 86, 35, 11, 83, 168, 109, 69, 53, 21, 189, 225, 61, 89, 76, 221, 211, 36, 241, 161, 124, 84, 243, 171, 216, 91, 84, 102, 160, 226, 223, 56, 145, 209, 140, 9, 84, 47, 145, 220, 214, 125, 63, 218, 32, 211, 172, 87, 82, 132, 128, 223, 126, 215, 13, 3, 72, 212, 5, 107, 197, 68, 183, 33, 127, 191, 218, 1, 3, 211, 115, 2, 162, 35, 99, 148, 182, 43, 187, 3, 62, 239, 125, 252, 93, 116, 253, 186, 126, 218, 201, 212, 21, 183, 15, 3, 225, 127, 44, 96, 224, 76, 254, 8, 225, 177, 26, 31, 190, 28, 159, 5, 6, 69, 113, 99, 209, 65, 50, 225, 127, 10, 58, 255, 131, 108, 135, 185, 1, 123, 232, 212, 183, 137, 8, 124, 139, 92, 52, 68, 113, 16, 70, 73, 180, 191, 252, 234, 90, 227, 231, 64, 159, 192, 63, 68, 0, 109, 205, 5, 33, 148, 189, 33, 239, 230, 69, 197, 193, 238, 97, 227, 180, 221, 149, 107, 126, 213, 173, 178, 113, 10, 112, 10, 44, 10, 75, 130, 240, 157, 228, 91, 34, 181, 110, 143, 212, 54, 47, 158, 41, 201, 181, 236, 231, 238, 193, 243, 190, 142, 195, 161, 89, 219, 36, 127, 178, 170, 198, 43, 171, 10, 66, 120, 154, 223, 160, 77, 204, 168, 228, 69, 134, 195, 84, 36, 24, 209, 37, 180, 71, 208, 116, 69, 239, 123, 28, 220, 112, 24, 43, 122, 15, 110, 61, 245, 112, 148, 53, 79, 116, 188, 65, 176, 152, 49, 105, 159, 69, 250, 85, 235, 131, 51, 221, 81, 79, 6, 93, 74, 8, 210, 126, 85, 120, 46, 37, 83, 100, 54, 89, 30, 50, 70, 175, 124, 182, 205, 216, 148, 175, 58, 35, 214, 184, 238, 163, 86, 234, 109, 196, 208, 53, 39, 155, 44, 34, 58, 125, 217, 44, 111, 21, 247, 146, 255, 26, 251, 254, 222, 131, 192, 158, 7, 132, 5, 18, 59, 46, 178, 14, 92, 216, 95, 245, 186, 162, 53, 140, 253, 131, 121, 99, 79, 252, 33, 195, 37, 113, 53, 233, 138, 8, 82, 124, 67, 138, 2, 97, 211, 111, 135, 248, 65, 123, 242, 127, 242, 6, 45, 3, 79, 144, 83, 186, 187, 7, 246, 219, 115, 234, 134, 201, 115, 93, 109, 12, 58, 237, 146, 243, 35, 123, 14, 168, 3, 185, 70, 42, 209, 128, 237, 99, 184, 206, 175, 36, 184, 231, 239, 233, 247, 74, 99, 94, 131, 7, 188, 170, 158, 250, 104, 122, 139, 33, 115, 51, 157, 251, 37, 56, 96, 87, 114, 13, 157, 81, 29, 1, 242, 51, 246, 77, 242, 163, 126, 100, 154, 145, 148, 168, 241, 28, 52, 224, 6, 71, 168, 214, 34, 251, 87, 65, 164, 77, 252, 47, 110, 35, 106, 142, 240, 243, 246, 30, 88, 163, 70, 71, 178, 231, 13, 70, 23, 29, 68, 234, 172, 101, 242, 235, 187, 168, 156, 126, 226, 165, 117, 224, 10, 126, 217, 27, 250, 88, 174, 180, 54, 172, 116, 120, 121, 17, 69, 243, 140, 146, 52, 25, 111, 160, 240, 247, 192, 142, 92, 238, 19, 60, 69, 253, 91, 238, 233, 250, 96, 204, 35, 25, 145, 195, 245, 211, 215, 50, 22, 88, 209, 139, 64, 253, 122, 251, 172, 170, 139, 65, 240, 151, 126, 203, 98, 67, 145, 247, 16, 61, 136, 114, 106, 15, 81, 214, 65, 115, 223, 7, 62, 137, 79, 22, 144, 41, 8, 239, 9, 184, 132, 66, 140, 185, 102, 45, 94, 163, 112, 36, 77, 14, 76, 228, 242, 168, 61, 49, 207, 141, 134, 151, 55, 210, 228, 150, 230, 163, 203, 231, 219, 73, 123, 154, 91, 199, 204, 19, 194, 213, 15, 187, 151, 222, 202, 229, 41, 22, 201, 143, 103, 197, 3, 6, 175, 73, 222, 25, 153, 5, 74, 22, 96, 85, 20, 212, 27, 228, 15, 19, 154, 27, 91, 241, 107, 70, 92, 160, 149, 5, 230, 126, 242, 243, 176, 155, 115, 81, 132, 8, 51, 67, 64, 2, 106, 194, 181, 240, 93, 157, 181, 21, 172, 143, 174, 180, 128, 44, 96, 147, 161, 145, 118, 166, 20, 232, 20, 160, 238, 135, 229, 201, 103, 149, 232, 11, 68, 203, 99, 150, 172, 248, 31, 48, 172, 151, 209, 85, 56, 15, 170, 243, 65, 219, 12, 142, 95, 51, 20, 6, 217, 242, 148, 218, 41, 25, 97, 247, 75, 61, 192, 194, 36, 118, 137, 112, 90, 40, 156, 206, 134, 87, 79, 156, 244, 17, 106, 31, 172, 133, 51, 212, 165, 112, 196, 182, 178, 154, 239, 64, 85, 152, 6, 27, 115, 83, 141, 152, 162, 199, 131, 233, 198, 74, 222, 11, 30, 109, 93, 108, 162, 133, 178, 81, 179, 210, 86, 204, 47, 46, 171, 236, 99, 254, 68, 163, 235, 29, 9, 73, 165, 236, 160, 182, 141, 246, 23, 154, 29, 7, 94, 223, 238, 76, 0, 149, 4, 20, 195, 186, 115, 97, 213, 217, 156, 251, 103, 8, 91, 108, 230, 58, 33, 50, 90, 7, 246, 98, 68, 178, 249, 106, 255, 68, 113, 224, 98, 195, 36, 45, 252, 81, 107, 71, 19, 202, 35, 34, 83, 159, 228, 194, 194, 88, 245, 182, 246, 144, 15, 69, 228, 139, 118, 233, 24, 90, 19, 82, 109, 240, 195, 234, 14, 42, 119, 56, 205, 197, 121, 120, 157, 124, 193, 67, 49, 241, 244, 54, 229, 53, 118, 56, 54, 230, 165, 124, 37, 201, 234, 138, 201, 206, 231, 97, 247, 192, 246, 67, 13, 192, 224, 66, 80, 216, 145, 103, 198, 165, 174, 119, 101, 211, 180, 163, 168, 10, 68, 19, 244, 250, 249, 12, 197, 185, 144, 87, 93, 252, 235, 180, 151, 225, 84, 131, 232, 177, 254, 91, 253, 63, 193, 204, 4, 60, 80, 95, 43, 129, 143, 186, 185, 79, 21, 146, 5, 171, 206, 122, 144, 177, 23, 185, 99, 51, 9, 59, 0, 199, 249, 29, 128, 68, 3, 21, 192, 139, 10, 51, 68, 203, 66, 206, 22, 85, 106, 228, 181, 213, 120, 234, 49, 168, 33, 154, 236, 239, 58, 213, 205, 250, 237, 216, 102, 198, 11, 49, 153, 188, 70, 39, 167, 187, 244, 1, 145, 228, 95, 151, 42, 3, 188, 131, 148, 233, 9, 118, 61, 40, 138, 179, 118, 143, 106, 159, 111, 158, 58, 157, 200, 128, 63, 137, 98, 215, 67, 159, 70, 250, 78, 147, 15, 4, 166, 228, 178, 13, 204, 215, 233, 74, 130, 206, 228, 242, 26, 161, 178, 169, 135, 180, 111, 122, 71, 85, 237, 34, 151, 21, 41, 99, 214, 217, 157, 187, 57, 176, 61, 3, 193, 178, 255, 204, 254, 226, 160, 231, 139, 251, 121, 72, 53, 14, 159, 80, 41, 177, 169, 114, 11, 38, 102, 156, 173, 40, 11, 125, 180, 15, 165, 168, 129, 248, 254, 3, 94, 186, 44, 160, 19, 91, 46, 61, 239, 95, 255, 2, 232, 136, 59, 116, 207, 41, 167, 78, 209, 195, 81, 211, 161, 34, 119, 121, 25, 48, 220, 91, 60, 203, 65, 105, 165, 54, 33, 72, 18, 201, 178, 49, 82, 138, 229, 251, 177, 11, 96, 103, 64, 143, 118, 68, 31, 57, 132, 209, 199, 136, 146, 14, 229, 65, 98, 84, 217, 122, 25, 158, 130, 142, 75, 222, 224, 175, 59, 125, 200, 243, 1, 50, 56, 73, 251, 52, 3, 203, 173, 20, 196, 109, 33, 22, 28, 150, 47, 252, 121, 156, 201, 158, 85, 149, 184, 122, 6, 156, 120, 172, 252, 157, 51, 240, 189, 8, 121, 10, 254, 164, 102, 217, 130, 89, 208, 180, 218, 198, 70, 114, 43, 181, 139, 142, 46, 10, 178, 152, 198, 201, 226, 3, 30, 159, 165, 140, 152, 68, 193, 167, 92, 38, 11, 143, 186, 10, 181, 83, 66, 104, 174, 30, 42, 33, 63, 55, 112, 106, 221, 187, 202, 120, 245, 227, 55, 35, 188, 114, 190, 239, 29, 190, 226, 122, 187, 52, 118, 163, 46, 147, 255, 44, 238, 12, 212, 49, 246, 225, 206, 157, 224, 190, 192, 127, 1, 34, 152, 197, 205, 15, 165, 15, 251, 58, 181, 48, 14, 179, 221, 250, 31, 190, 97, 122, 92, 166, 36, 125, 23, 177, 45, 53, 16, 45, 161, 205, 226, 193, 34, 116, 46, 54, 43, 138, 81, 63, 232, 216, 67, 212, 180, 150, 155, 48, 27, 39, 206, 119, 242, 216, 73, 179, 19, 186, 239, 105, 213, 134, 29, 47, 67, 189, 86, 176, 51, 46, 155, 176, 173, 136, 8, 229, 81, 227, 237, 16, 179, 59, 92, 110, 70, 25, 94, 126, 113, 38, 203, 129, 92, 107, 223, 40, 225, 251, 106, 4, 21, 18, 168, 236, 13, 143, 25, 148, 238, 6, 24, 71, 228, 103, 95, 228, 5, 41, 66, 197, 104, 14, 238, 20, 152, 116, 119, 253, 109, 132, 165, 178, 151, 191, 164, 156, 57, 172, 131, 10, 155, 113, 159, 41, 190, 100, 215, 94, 175, 49, 47, 154, 207, 212, 186, 36, 146, 195, 80, 0, 107, 121, 40, 112, 31, 159, 121, 70, 88, 218, 113, 139, 0, 157, 135, 7, 234, 95, 212, 75, 3, 250, 177, 103, 83, 153, 90, 26, 55, 21, 153, 241, 84, 26, 168, 119, 157, 123, 170, 16, 224, 106, 117, 1, 42, 56, 71, 19, 161, 106, 27, 129, 104, 107, 154, 222, 152, 11, 95, 221, 6, 167, 42, 121, 48, 130, 5, 98, 6, 9, 42, 134, 72, 134, 247, 13, 1, 7, 1, 160, 130, 5, 83, 4, 130, 5, 79, 48, 130, 5, 75, 48, 130, 5, 71, 6, 11, 42, 134, 72, 134, 247, 13, 1, 12, 10, 1, 2, 160, 130, 4, 238, 48, 130, 4, 234, 48, 28, 6, 10, 42, 134, 72, 134, 247, 13, 1, 12, 1, 3, 48, 14, 4, 8, 34, 116, 121, 209, 47, 27, 82, 130, 2, 2, 8, 0, 4, 130, 4, 200, 101, 10, 91, 45, 187, 141, 27, 192, 239, 243, 101, 206, 57, 201, 190, 219, 183, 161, 207, 219, 252, 199, 12, 185, 249, 32, 90, 106, 80, 144, 236, 96, 29, 246, 142, 43, 142, 128, 52, 241, 216, 172, 113, 64, 127, 251, 177, 159, 24, 136, 125, 189, 244, 94, 25, 253, 82, 122, 108, 240, 190, 184, 214, 200, 40, 34, 172, 14, 128, 209, 121, 58, 200, 225, 177, 170, 103, 207, 213, 8, 194, 173, 255, 177, 64, 223, 66, 75, 36, 93, 126, 239, 26, 61, 142, 153, 192, 17, 170, 160, 126, 126, 215, 97, 0, 186, 252, 178, 220, 241, 27, 224, 16, 82, 119, 5, 131, 202, 105, 162, 156, 187, 170, 128, 96, 67, 99, 36, 187, 212, 37, 9, 82, 70, 20, 24, 204, 148, 112, 53, 33, 141, 183, 103, 87, 95, 147, 223, 68, 210, 100, 18, 177, 35, 18, 6, 100, 179, 2, 24, 106, 175, 118, 122, 227, 235, 117, 31, 31, 237, 227, 51, 43, 138, 242, 135, 52, 221, 132, 82, 29, 66, 16, 93, 196, 54, 122, 165, 151, 32, 110, 213, 130, 36, 151, 107, 198, 117, 154, 180, 142, 83, 222, 154, 143, 0, 5, 89, 117, 235, 93, 187, 4, 166, 220, 28, 96, 28, 190, 70, 167, 29, 125, 12, 116, 154, 176, 83, 133, 189, 0, 227, 6, 188, 89, 207, 183, 40, 238, 254, 214, 239, 251, 42, 193, 142, 120, 121, 34, 28, 72, 207, 32, 38, 239, 19, 73, 217, 20, 150, 31, 182, 200, 254, 36, 32, 131, 254, 231, 111, 113, 28, 110, 109, 48, 101, 74, 217, 252, 171, 196, 181, 54, 214, 208, 206, 33, 87, 162, 18, 50, 77, 99, 234, 45, 81, 136, 122, 237, 24, 23, 223, 188, 68, 89, 208, 129, 192, 194, 4, 60, 42, 149, 197, 182, 4, 170, 196, 212, 32, 166, 220, 1, 227, 46, 203, 45, 78, 33, 159, 79, 237, 13, 179, 63, 217, 246, 67, 144, 30, 32, 214, 44, 193, 64, 214, 154, 254, 136, 186, 100, 37, 248, 223, 170, 104, 104, 173, 243, 130, 216, 166, 49, 38, 95, 159, 25, 22, 129, 150, 244, 75, 115, 108, 234, 87, 223, 85, 93, 228, 162, 145, 18, 102, 53, 174, 142, 89, 208, 158, 232, 74, 104, 129, 185, 182, 192, 11, 180, 207, 29, 198, 104, 104, 76, 17, 70, 116, 31, 216, 87, 39, 59, 36, 53, 140, 82, 90, 66, 252, 185, 93, 94, 175, 136, 228, 67, 166, 187, 20, 87, 245, 249, 195, 199, 116, 110, 158, 44, 139, 85, 17, 98, 93, 177, 68, 25, 0, 174, 213, 14, 82, 142, 2, 12, 21, 80, 234, 104, 44, 220, 222, 40, 156, 120, 204, 29, 109, 101, 195, 204, 97, 178, 45, 87, 234, 192, 135, 153, 182, 66, 52, 45, 240, 232, 51, 97, 244, 10, 24, 16, 194, 159, 88, 188, 120, 26, 112, 52, 178, 216, 7, 188, 89, 120, 112, 188, 245, 10, 76, 197, 244, 248, 155, 158, 149, 128, 129, 65, 100, 42, 183, 145, 134, 59, 29, 249, 232, 248, 109, 70, 53, 122, 160, 129, 144, 49, 121, 150, 200, 182, 228, 233, 27, 199, 125, 237, 45, 205, 110, 225, 1, 47, 230, 212, 185, 120, 1, 19, 184, 143, 76, 109, 36, 216, 125, 75, 139, 11, 230, 227, 4, 226, 211, 121, 5, 104, 102, 73, 79, 253, 205, 10, 206, 243, 8, 3, 244, 79, 126, 156, 121, 242, 48, 200, 36, 11, 17, 107, 52, 37, 52, 228, 201, 162, 33, 250, 9, 193, 22, 15, 236, 182, 126, 118, 35, 210, 160, 201, 219, 132, 102, 32, 26, 239, 122, 72, 45, 29, 213, 145, 154, 94, 5, 248, 242, 77, 83, 203, 83, 246, 122, 167, 221, 252, 213, 198, 149, 28, 10, 236, 155, 85, 85, 9, 93, 9, 25, 146, 70, 217, 247, 162, 212, 32, 2, 64, 113, 15, 125, 140, 140, 42, 101, 98, 57, 45, 210, 238, 148, 209, 81, 255, 220, 26, 225, 45, 57, 80, 168, 238, 236, 8, 16, 134, 21, 237, 16, 242, 58, 145, 156, 124, 124, 0, 138, 76, 5, 165, 52, 239, 197, 197, 50, 251, 93, 85, 20, 198, 117, 183, 253, 228, 142, 74, 159, 237, 50, 251, 160, 198, 226, 26, 110, 191, 178, 204, 242, 215, 91, 88, 162, 244, 230, 80, 101, 123, 66, 53, 192, 106, 233, 50, 171, 16, 98, 151, 149, 188, 175, 64, 194, 29, 65, 113, 166, 158, 21, 198, 75, 199, 86, 208, 150, 141, 164, 250, 168, 188, 22, 156, 101, 241, 148, 120, 129, 132, 127, 164, 123, 118, 188, 70, 212, 164, 9, 140, 143, 244, 122, 244, 26, 0, 222, 96, 135, 202, 133, 193, 95, 201, 244, 76, 140, 145, 0, 134, 22, 74, 191, 218, 192, 45, 234, 92, 205, 86, 208, 153, 240, 74, 57, 51, 169, 75, 211, 145, 91, 29, 237, 23, 228, 152, 178, 242, 127, 4, 38, 82, 218, 0, 144, 43, 242, 125, 152, 239, 244, 150, 196, 84, 19, 12, 133, 56, 76, 124, 27, 99, 142, 85, 186, 202, 6, 77, 59, 185, 66, 206, 10, 63, 198, 27, 157, 50, 100, 148, 81, 158, 65, 206, 12, 163, 19, 82, 223, 223, 166, 218, 185, 179, 66, 221, 53, 92, 5, 44, 44, 91, 68, 157, 102, 104, 248, 231, 220, 123, 54, 114, 94, 237, 56, 171, 198, 91, 211, 92, 205, 184, 45, 228, 172, 252, 155, 8, 36, 31, 214, 165, 218, 134, 174, 172, 70, 116, 103, 111, 237, 138, 169, 72, 223, 251, 65, 184, 209, 248, 161, 161, 109, 224, 112, 172, 213, 26, 154, 14, 114, 235, 162, 182, 92, 137, 222, 44, 14, 83, 239, 97, 168, 137, 95, 80, 231, 184, 167, 65, 26, 136, 130, 40, 117, 233, 4, 145, 174, 49, 82, 61, 218, 204, 42, 143, 84, 255, 195, 255, 46, 252, 116, 69, 35, 102, 211, 67, 15, 81, 33, 44, 158, 231, 181, 188, 95, 85, 123, 234, 85, 15, 188, 168, 219, 200, 29, 33, 149, 154, 207, 41, 95, 127, 99, 121, 100, 20, 201, 221, 141, 8, 220, 4, 132, 191, 193, 189, 63, 50, 81, 134, 8, 173, 49, 63, 90, 218, 17, 195, 69, 140, 17, 141, 123, 123, 134, 49, 78, 49, 156, 222, 123, 245, 253, 88, 177, 77, 206, 26, 15, 19, 57, 188, 135, 32, 91, 72, 35, 180, 183, 16, 155, 115, 217, 90, 108, 196, 194, 227, 108, 59, 126, 181, 177, 148, 226, 65, 99, 47, 230, 249, 14, 201, 104, 155, 191, 158, 145, 5, 109, 88, 125, 186, 145, 77, 15, 214, 149, 217, 110, 208, 107, 72, 172, 35, 2, 12, 236, 245, 225, 135, 160, 69, 219, 249, 50, 24, 241, 24, 121, 217, 38, 105, 139, 212, 233, 135, 187, 252, 182, 84, 141, 205, 16, 232, 97, 168, 216, 201, 105, 209, 210, 119, 213, 161, 184, 60, 137, 103, 112, 94, 140, 68, 49, 196, 195, 235, 105, 232, 168, 173, 142, 178, 217, 107, 19, 218, 91, 100, 98, 39, 218, 121, 235, 87, 187, 217, 0, 128, 19, 216, 227, 246, 254, 53, 133, 49, 70, 48, 31, 6, 9, 42, 134, 72, 134, 247, 13, 1, 9, 20, 49, 18, 30, 16, 0, 66, 0, 97, 0, 114, 0, 107, 0, 80, 0, 117, 0, 115, 0, 104, 48, 35, 6, 9, 42, 134, 72, 134, 247, 13, 1, 9, 21, 49, 22, 4, 20, 206, 241, 31, 180, 60, 189, 54, 61, 207, 254, 234, 236, 195, 214, 138, 23, 108, 207, 31, 97, 48, 48, 48, 33, 48, 9, 6, 5, 43, 14, 3, 2, 26, 5, 0, 4, 20, 230, 122, 192, 87, 197, 72, 184, 96, 246, 56, 149, 242, 37, 165, 29, 4, 89, 180, 108, 146, 4, 8, 234, 222, 82, 13, 98, 36, 189, 233, 2, 1, 1}
	}
	//线上证书
	return []byte{}
}

func postPush(category string, title string, body string, deviceToken string,  params map[string]interface{}) error{

	notification := &apns2.Notification{}
	notification.DeviceToken = deviceToken

	payload := payload.NewPayload().Sound("1107").Category("myNotificationCategory")
	badge := params["badge"]
	if badge != nil {
		badgeStr, pass := badge.(string)
		if pass {
			badgeNum, err := strconv.Atoi(badgeStr)
			if err == nil {
				payload = payload.Badge(badgeNum)
			}
		}
	}

	for key, value := range params {
		payload = payload.Custom(key, value)
	}
	if len(title) > 0 {
		payload.AlertTitle(title)
	}
	if len(body) > 0 {
		payload.AlertBody(body)
	}
	notification.Payload = payload
	notification.Topic = "me.fin.bark"
	res, err := apnsClient.Push(notification)

	if err != nil {
		log.Println("Error:", err)
		return errors.New("与苹果推送服务器传输数据失败")
	}
	log.Printf("%v %v %v\n", res.StatusCode, res.ApnsID, res.Reason)
	if res.StatusCode == 200 {
		return nil
	}else{
		return errors.New("推送发送失败 " + res.Reason)
	}


}

var IsDev bool = false
var boltDB *bolt.DB
var apnsClient *apns2.Client
func main()  {
	//f,_:= os.Open("./BarkPush.p12")
	//t,_ := ioutil.ReadAll(f)
	//
	//str := ""
	//for _,val := range t {
	//	str += ", "
	//	str += strconv.Itoa(int(val))
	//}
	//
	//fmt.Printf(string(t))
	ip := flag.String("ip",  "0.0.0.0", "http listen ip")
	port := flag.Int("port",  8080, "http listen port")
	dev := flag.Bool("dev", false, "develop推送，请忽略此参数，设置此参数为True会导致推送失败")
	flag.Parse()

	IsDev = *dev

	db, err := bolt.Open("bark.db", 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer  db.Close()
	boltDB = db

	boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("device"))
		if err != nil {
			log.Fatalln(err)
		}
		return err
	})

	cert, err := certificate.FromP12Bytes(getb(),"bp")
	if err != nil {
		log.Fatalln("cer error")
	}
	apnsClient = apns2.NewClient(cert).Production()



	addr := *ip + ":" + strconv.Itoa(*port)
	log.Println("Serving HTTP on " + addr)

	r := bone.New()
	r.Get("/ping", http.HandlerFunc(ping))
	r.Post("/ping", http.HandlerFunc(ping))

	r.Get("/register", http.HandlerFunc(register))
	r.Post("/register", http.HandlerFunc(register))

	r.Get("/:key/:body", http.HandlerFunc(Index))
	r.Post("/:key/:body", http.HandlerFunc(Index))

	r.Get("/:key/:title/:body", http.HandlerFunc(Index))
	r.Post("/:key/:title/:body", http.HandlerFunc(Index))

	r.Get("/:key/:category/:title/:body", http.HandlerFunc(Index))
	r.Post("/:key/:category/:title/:body", http.HandlerFunc(Index))


	log.Fatal(http.ListenAndServe(addr, r))
}

