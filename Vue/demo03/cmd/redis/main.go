package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"gopkg.in/mgo.v2"
)

type DBInstanceConfig struct {
	ZoneID          string                 `json:"zoneID"`
	NetworkTypes    string                 `json:"networkTypes"`
	regionID        string                 `json:"regionID"`
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

func main() {

	client, err := rds.NewClientWithAccessKey("cn-beijing", os.Getenv("AK"), os.Getenv("SK"))

	request := rds.CreateDescribeAvailableResourceRequest()
	request.Scheme = "https"

	// required
	request.InstanceChargeType = "PostPaid"

	response, err := client.DescribeAvailableResource(request)
	if err != nil {
		fmt.Print(err.Error())
	}

	sess, err := getMSession("mongodb://10.0.1.206/alicloud")
	if err != nil {
		panic(err)
	}
	defer sess.Close()
	// Optional. Switch the session to a monotonic behavior.
	sess.SetMode(mgo.Monotonic, true)

	// get c
	c := sess.DB("alicloud").C("rds")

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
								regionID:        regionID,
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

	// fmt.Printf("response is %#v\n", response)
}
