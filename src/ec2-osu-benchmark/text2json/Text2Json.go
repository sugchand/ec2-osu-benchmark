package text2json

import (
    "os"
    "strconv"
    "time"
    "encoding/json"
    "io/ioutil"
    "strings"
    "bufio"
    "ec2-osu-benchmark/logging"
    "ec2-osu-benchmark/errors"
    "ec2-osu-benchmark/config"
)

//convert BW, latency txt file contents to json file
// The text file would be in the following format.
// # Size      Bandwidth (MB/s)
//   1                       6.09
//   2                      15.37
//   4                      35.10
//   8                      55.43
//
//  Will add the timestamp when the json file is created.
//
// JSON file format would be
// {
//     "osu-bw": {
//         "timestamp": 1559324795,
//         "values": {
//                     "bw"      : [234, 456, 678, ....],
//                     "pktsize" : [1, 8, 128, 256, ...],
//                   },
//                }
//}

type OsuBWTuple struct {
    Bw      float64 `json:"bw"`
    Pktsize int `json:"pktsize"`
}
type OsuBW []OsuBWTuple

type OsuBiBW struct {
    Bw      []float64 `json:"bw"`
    Pktsize []int `json:"pktsize"`
}

type OsuLatencyTuple struct {
    Latency float64 `json:"latency"`
    Pktsize      int `json:"pktsize"`
}

type OsuLatency []OsuLatencyTuple

type OSUResults struct {
    Timestamp time.Time `json:"timestamp"`
    OsuBW `json:"OsuBW"`
    OsuBiBW `json:"OsuBiBW"`
    OsuLatency `json:"OsuLatency"`
}

type Text2Json struct {
    jsonFile string
    filelist []string
    jsonResults *OSUResults
}

func (txt2jsonObj *Text2Json)GetAllFiles(path string) error {
    logger := logging.GetLoggerInstance()
    fileNames, err := ioutil.ReadDir(path)
    if err != nil {
        logger.Error("Failed to get all the results files in %s", path)
        return err
    }
    txt2jsonObj.filelist = make([]string, len(fileNames))
    for idx, f := range fileNames {
        txt2jsonObj.filelist[idx] = path + f.Name()
    }
    return errors.OP_SUCCESS
}

func (txt2jsonObj *Text2Json)IsLatencyFile(fileName string) bool {
    return strings.Contains(fileName, "latency") ||
            strings.Contains(fileName, "LATENCY")
}

func (txt2jsonObj *Text2Json)ReadOSULatencyFile(fileName string,
                            latencyResults *OsuLatency) error {
    logger := logging.GetLoggerInstance()    
    fileName = strings.Trim(fileName, "\n")

    file, err := os.Open(fileName)
    if err != nil {
        logger.Error("Failed to open file %s", fileName)
        return err
    }

    defer file.Close()
    latencyTupleSet := make([]OsuLatencyTuple, 0)
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "#") {
            // Comment line, continue to next line
            continue
        }
        lineArr := strings.Fields(line)
        if len(lineArr) < 2 {
            // Invalid data, cannot copy
            continue
        }
        var latTuple OsuLatencyTuple
        latTuple.Pktsize, _ = strconv.Atoi(lineArr[0])
        latTuple.Latency, _ = strconv.ParseFloat(lineArr[1], 64)
        latencyTupleSet =
                    append(latencyTupleSet, latTuple)
    }
    *latencyResults = latencyTupleSet
    return errors.OP_SUCCESS
}

func (txt2jsonObj *Text2Json)Init(resPath string) {
    txt2jsonObj.GetAllFiles(resPath)
    txt2jsonObj.jsonResults = new(OSUResults)
    txt2jsonObj.jsonFile = resPath + "osu-report.json"
}

func (txt2jsonObj *Text2Json)WriteLatency2Json() {
    logger := logging.GetLoggerInstance()
    var latencyResults OsuLatency
    for _, fileName := range txt2jsonObj.filelist {
        if txt2jsonObj.IsLatencyFile(fileName) {
            // Process only latency files
            txt2jsonObj.ReadOSULatencyFile(fileName, &latencyResults)
            txt2jsonObj.jsonResults.OsuLatency = latencyResults
            logger.Info("Processing of latency is complete")
            break
        }
    }
}

func (txt2jsonObj *Text2Json)WriteTimestamp() error {
    var err error
    file_set := strings.Split(txt2jsonObj.jsonFile, "/")
    timestr := file_set[len(file_set) - 3]
    txt2jsonObj.jsonResults.Timestamp, err = time.Parse(config.DEFAULT_TIME_LAYOUT,
                                        timestr)
    if err == nil {
        return errors.OP_SUCCESS
    }
    return err
}

// Function to write the structure to json result file.
func (txt2jsonObj *Text2Json)WriteJsonFile() error{
    logger := logging.GetLoggerInstance()
    jsonBytes, err := json.MarshalIndent(txt2jsonObj.jsonResults, "", "  ")
    if err != nil {
        logger.Error("Failedto marshal json file, cannot write.. \n")
        return err
    }
    err = ioutil.WriteFile(txt2jsonObj.jsonFile, jsonBytes, 0644)
    if err == nil {
        return errors.OP_SUCCESS
    }
    return err
}

func (txt2jsonObj *Text2Json)ProcessResults2Json() error {
    txt2jsonObj.WriteTimestamp()
    txt2jsonObj.WriteLatency2Json()
    err := txt2jsonObj.WriteJsonFile()
    return err
}

