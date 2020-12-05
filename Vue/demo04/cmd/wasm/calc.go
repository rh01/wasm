package main // import "calc.wasm"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"syscall/js"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var jb = js.Global()

type DBInstanceConfig struct {
	ZoneID          string                 `json:"zoneID"`
	NetworkTypes    string                 `json:"networkTypes"`
	RegionID        string                 `json:"regionID"`
	ZoneStatue      string                 `json:"zoneStatue"`
	Engine          string                 `json:"engine"`
	EngineVersion   string                 `json:"engineVersion"`
	Category        string                 `json:"category"`
	StorageType     string                 `json:"storageType"`
	DBInstanceClass string                 `json:"dbInstanceClass"`
	DBInstanceRange map[string]interface{} `json:"dbInstanceRange"`
	StorageRange    string                 `json:"storageRange"`
}

func dumpMap(space string, m map[string]interface{}) {
	for k, v := range m {
		if mv, ok := v.(map[string]interface{}); ok {
			fmt.Printf("{ \"%v\": \n", k)
			dumpMap(space+"\t", mv)
			fmt.Printf("}\n")
		} else {
			fmt.Printf("%v %v : %v\n", space, k, v)
		}
	}
}

func getMSession(conn string) (*mgo.Session, error) {
	return mgo.Dial(conn)
}

func sync() {
	// set up mongodb client
	sess, err := getMSession("mongodb://localhost/matrix")
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	sess.Clone()

	// Optional. Switch the session to a monotonic behavior.
	sess.SetMode(mgo.Monotonic, true)

	sess.DB("test").C("rds").Find(nil).Count()
	// set up an aliyun rds client
	client, err := rds.NewClientWithAccessKey("cn-beijing", os.Getenv("AK"), os.Getenv("SK"))

	request := rds.CreateDescribeAvailableResourceRequest()
	request.Scheme = "https"

	// required
	request.InstanceChargeType = "PostPaid"

	response, err := client.DescribeAvailableResource(request)
	if err != nil {
		fmt.Print(err.Error())
	}

	// get c
	c := sess.DB("matrix").C("rds")

	for _, zone := range response.AvailableZones.AvailableZone {
		zoneID := zone.ZoneId
		networktype := zone.NetworkTypes
		regionID := zone.RegionId
		zoneStatue := zone.Status

		for _, engine := range zone.SupportedEngines.SupportedEngine {
			engineName := engine.Engine
			for _, supportEngine := range engine.SupportedEngineVersions.SupportedEngineVersion {
				supportEngineVersion := supportEngine.Version
				for _, supportCategory := range supportEngine.SupportedCategorys.SupportedCategory {
					category := supportCategory.Category
					for _, supportedStorageType := range supportCategory.SupportedStorageTypes.SupportedStorageType {
						storageType := supportedStorageType.StorageType
						for _, availableResources := range supportedStorageType.AvailableResources.AvailableResource {
							dbInstanceClass := availableResources.DBInstanceClass
							var buf bytes.Buffer
							err := json.NewEncoder(&buf).Encode(&availableResources.DBInstanceStorageRange)
							if err != nil {
								continue
							}
							dbInstanceRange := make(map[string]interface{})
							err = json.Unmarshal(buf.Bytes(), &dbInstanceRange)
							if err != nil {
								continue
							}
							storageRange := availableResources.StorageRange
							config := &DBInstanceConfig{
								ZoneID:          zoneID,
								NetworkTypes:    networktype,
								RegionID:        regionID,
								ZoneStatue:      zoneStatue,
								Engine:          engineName,
								EngineVersion:   supportEngineVersion,
								Category:        category,
								StorageType:     storageType,
								DBInstanceClass: dbInstanceClass,
								DBInstanceRange: dbInstanceRange,
								StorageRange:    storageRange,
							}
							err = c.Insert(config)
							if err != nil {
								log.Fatal(err)
							}
						}
					}
				}
			}
		}
	}
}

func syncAll(this js.Value, args []js.Value) interface{} {
	// Handler for the Promise
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {

			// set up an aliyun rds client
			client, err := rds.NewClientWithAccessKey("cn-beijing", "", "")
			// client.InitClientConfig().HttpTransport.ProxyConnectHeader.Add(
			// 	"Referer", "http://cs.aliyuncs.com",	
			// )
			client.SetHttpsProxy("http://cs.aliyuncs.com")

			request := rds.CreateDescribeAvailableResourceRequest()
			request.Scheme = "https"
			
			// request.AcceptFormat = "json"
			// request.SetReadTimeout = 5 * time.Minute
			request.SetReadTimeout(10 * time.Second)             
			// request.Domain = "cs.aliyuncs.com"
			// SetBrowserRequestMode
			request.SetDomain("cs.aliyuncs.com")
			// required
			request.InstanceChargeType = "PostPaid"
			resp, err := client.DescribeAvailableResource(request)
			if err != nil {
				fmt.Print(err.Error())
			}

			var configs = make([]DBInstanceConfig, 0)
			for _, zone := range resp.AvailableZones.AvailableZone {
				zoneID := zone.ZoneId
				networktype := zone.NetworkTypes
				regionID := zone.RegionId
				zoneStatue := zone.Status

				for _, engine := range zone.SupportedEngines.SupportedEngine {
					engineName := engine.Engine
					for _, supportEngine := range engine.SupportedEngineVersions.SupportedEngineVersion {
						supportEngineVersion := supportEngine.Version
						for _, supportCategory := range supportEngine.SupportedCategorys.SupportedCategory {
							category := supportCategory.Category
							for _, supportedStorageType := range supportCategory.SupportedStorageTypes.SupportedStorageType {
								storageType := supportedStorageType.StorageType
								for _, availableResources := range supportedStorageType.AvailableResources.AvailableResource {
									dbInstanceClass := availableResources.DBInstanceClass
									var buf bytes.Buffer
									err := json.NewEncoder(&buf).Encode(&availableResources.DBInstanceStorageRange)
									if err != nil {
										continue
									}
									dbInstanceRange := make(map[string]interface{})
									err = json.Unmarshal(buf.Bytes(), &dbInstanceRange)
									if err != nil {
										continue
									}
									storageRange := availableResources.StorageRange
									config := DBInstanceConfig{
										ZoneID:          zoneID,
										NetworkTypes:    networktype,
										RegionID:        regionID,
										ZoneStatue:      zoneStatue,
										Engine:          engineName,
										EngineVersion:   supportEngineVersion,
										Category:        category,
										StorageType:     storageType,
										DBInstanceClass: dbInstanceClass,
										DBInstanceRange: dbInstanceRange,
										StorageRange:    storageRange,
									}

									configs = append(configs, config)
								}
							}
						}
					}
				}
			}

			var buf bytes.Buffer
			err = json.NewEncoder(&buf).Encode(configs)
			if err != nil {
				errConstructor := js.Global().Get("Error")
				errObject := errConstructor.New(err.Error())
				reject.Invoke(errObject)
			}
			// The handler of a Promise doesn't return any value
			// "data" is a byte slice, so we need to convert it to a JS Uint8Array object
			arrayConstructor := js.Global().Get("Uint8Array")
			dataJS := arrayConstructor.New(len(configs))
			js.CopyBytesToJS(dataJS, buf.Bytes())

			// Create a Response object and pass the data
			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(dataJS)

			// Resolve the Promise
			resolve.Invoke(response)
		}()
		return nil
	})
	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

// demo01, async
func myGoFunc(this js.Value, args []js.Value) interface{} {
	// Handler for the Promise
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		// Run this code asynchronously
		go func() {
			// Cause a failure 50% of times
			if rand.Int()%2 == 0 {
				// Invoke the resolve function passing a plain JS object/dictionary
				resolve.Invoke(map[string]interface{}{
					"message": "Hooray, it worked!",
					"error":   nil,
				})
			} else {
				// Assume this were a Go error object
				err := errors.New("Nope, it failed")

				// Create a JS Error object and pass it to the reject function
				// The constructor for Error accepts a string,
				// so we need to get the error message as string from "err"
				errorConstructor := js.Global().Get("Error")
				errorObject := errorConstructor.New(err.Error())
				reject.Invoke(errorObject)
			}
		}()

		// The handler of a Promise doesn't return any value
		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

// demo02, async
func asyncOne(this js.Value, args []js.Value) interface{} {
	// Handler for the Promise: this is a JS function
	// It receives two arguments, which are JS functions themselves: resolve and reject
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		// Commented out because this Promise never fails
		// reject := args[1]

		// Now that we have a way to return the response to JS, spawn a goroutine
		// This way, we don't block the event loop and avoid a deadlock
		go func() {
			// Block the goroutine for 3 seconds
			time.Sleep(3 * time.Second)
			// Resolve the Promise, passing anything back to JavaScript
			// This is done by invoking the "resolve" function passed to the handler
			resolve.Invoke("Trentatré Trentini entrarono a Trento, tutti e trentatré trotterellando")
		}()

		// The handler of a Promise doesn't return any value
		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func fetchMongoDocument(this js.Value, args []js.Value) interface{} {
	// Get the URL as argument
	// args[0] is a js.Value, so we need to get a string out of it
	// connUrl := args[0].String()

	// Handler for the Promise
	// We need to return a Promise because HTTP requests are blocking in Go
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		// Run this code asynchronously
		go func() {
			// Make the mongo connection
			sess, err := mgo.Dial("127.0.0.1")
			defer sess.Close()

			sess.SetMode(mgo.Monotonic, true)

			c := sess.DB("alicloud").C("rds")
			var res = DBInstanceConfig{}

			q := bson.M{}
			err = c.Find(q).One(&res)

			if err != nil {
				errorConstructor := js.Global().Get("Error")
				errorObject := errorConstructor.New(err.Error())
				reject.Invoke(errorObject)
				return
			}

			// Marshal json format
			var buf bytes.Buffer
			err = json.NewEncoder(&buf).Encode(&res)
			if err != nil {
				// Handle errors here too
				errorConstructor := js.Global().Get("Error")
				errorObject := errorConstructor.New(err.Error())
				reject.Invoke(errorObject)
				return
			}

			data := buf.Bytes()
			// "data" is a byte slice, so we need to convert it to a JS Uint8Array object
			arrayConstructor := js.Global().Get("Uint8Array")
			dataJS := arrayConstructor.New(len(data))
			js.CopyBytesToJS(dataJS, data)

			// Create a Response object and pass the data
			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(dataJS)

			// Resolve the Promise
			resolve.Invoke(response)
		}()

		// The handler of a Promise doesn't return any value
		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func fetchHttp(this js.Value, args []js.Value) interface{} {
	// Get the URL as argument
	// args[0] is a js.Value, so we need to get a string out of it
	requestUrl := args[0].String()

	// Handler for the Promise
	// We need to return a Promise because HTTP requests are blocking in Go
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		// Run this code asynchronously
		go func() {
			// Make the HTTP request
			res, err := http.DefaultClient.Get(requestUrl)
			if err != nil {
				// Handle errors: reject the Promise if we have an error
				errorConstructor := js.Global().Get("Error")
				errorObject := errorConstructor.New(err.Error())
				reject.Invoke(errorObject)
				return
			}
			defer res.Body.Close()

			// Read the response body
			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				// Handle errors here too
				errorConstructor := js.Global().Get("Error")
				errorObject := errorConstructor.New(err.Error())
				reject.Invoke(errorObject)
				return
			}

			// "data" is a byte slice, so we need to convert it to a JS Uint8Array object
			arrayConstructor := js.Global().Get("Uint8Array")
			dataJS := arrayConstructor.New(len(data))
			js.CopyBytesToJS(dataJS, data)

			// Create a Response object and pass the data
			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(dataJS)

			// Resolve the Promise
			resolve.Invoke(response)
		}()

		// The handler of a Promise doesn't return any value
		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func add(this js.Value, i []js.Value) interface{} {
	value1 := i[0].String()
	value2 := i[1].String()
	int1, _ := strconv.Atoi(value1)
	int2, _ := strconv.Atoi(value2)

	if value1 == "" {
		int1 = 0
	}
	if value2 == "" {
		int2 = 0
	}

	result := int1 + int2

	return result
}

func sub(this js.Value, i []js.Value) interface{} {
	value1 := i[0].String()
	value2 := i[1].String()
	int1, _ := strconv.Atoi(value1)
	int2, _ := strconv.Atoi(value2)

	if value1 == "" {
		int1 = 0
	}
	if value2 == "" {
		int2 = 0
	}

	result := int1 - int2

	return result
}

func multi(this js.Value, i []js.Value) interface{} {
	value1 := i[0].String()
	value2 := i[1].String()
	int1, _ := strconv.Atoi(value1)
	int2, _ := strconv.Atoi(value2)

	if value1 == "" {
		int1 = 0
	}
	if value2 == "" {
		int2 = 0
	}

	result := int1 * int2

	return result
}

func divi(this js.Value, i []js.Value) interface{} {
	value1 := i[0].String()
	value2 := i[1].String()
	int1, _ := strconv.Atoi(value1)
	int2, _ := strconv.Atoi(value2)

	if value2 == "" || value2 == "0" {
		return "infinite"
	}

	result := int1 / int2

	return result
}

func main() {
	c := make(chan struct{}, 0)

	jb.Set("waAdd", js.FuncOf(add))
	jb.Set("waSub", js.FuncOf(sub))
	jb.Set("waMulti", js.FuncOf(multi))
	jb.Set("waDivi", js.FuncOf(divi))
	jb.Set("myGoFunc", js.FuncOf(myGoFunc))
	jb.Set("asyncOne", js.FuncOf(asyncOne))
	jb.Set("fetchHttp", js.FuncOf(fetchHttp))
	jb.Set("fetchMongoDocument", js.FuncOf(syncAll))

	println("Go Web Assembly Ready")

	<-c
}
