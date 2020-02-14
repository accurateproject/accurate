package utils

import "time"

type LoadInstance struct {
	LoadID           string `bson:"load_id"` // Unique identifier for the load
	RatingLoadID     string `bson:"rating_load_id"`
	AccountingLoadID string `bson:"accounting_load_id"`
	//TariffPlanID     string `bson:"tariff_plan_id"`    // Tariff plan identificator for the data loaded
	LoadTime time.Time `bson:"load_time"` // Time of load
}

type CacheFileInfo struct {
	Encoding string
	LoadInfo *LoadInstance
}

/*
func LoadCacheFileInfo(path string) (*CacheFileInfo, error) {
	// open data file
	dataFile, err := os.Open(filepath.Join(path, "cache.info"))
	defer dataFile.Close()
	if err != nil {
		Logger.Err("<cache decoder>: " + err.Error())
		return nil, err
	}

	filesInfo := &CacheFileInfo{}
	dataDecoder := json.NewDecoder(dataFile)
	err = dataDecoder.Decode(filesInfo)
	if err != nil {
		Logger.Err("<cache decoder>: " + err.Error())
		return nil, err
	}
	return filesInfo, nil
}

func SaveCacheFileInfo(path string, cfi *CacheFileInfo) error {
	if path == "" {
		return nil
	}
	// open data file
	// create a the path
	if err := os.MkdirAll(path, 0766); err != nil {
		Logger.Err("<cache encoder>:" + err.Error())
		return err
	}

	dataFile, err := os.Create(filepath.Join(path, "cache.info"))
	defer dataFile.Close()
	if err != nil {
		Logger.Err("<cache encoder>:" + err.Error())
		return err
	}

	// serialize the data
	dataEncoder := json.NewEncoder(dataFile)
	if err := dataEncoder.Encode(cfi); err != nil {
		Logger.Err("<cache encoder>:" + err.Error())
		return err
	}
	return nil
}

func CacheFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	return false
}
*/
