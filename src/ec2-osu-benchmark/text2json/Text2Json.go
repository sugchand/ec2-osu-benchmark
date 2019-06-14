package text2json

import (
    "bufio"
    "ec2-osu-benchmark/config"
    "ec2-osu-benchmark/errors"
    "ec2-osu-benchmark/logging"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
    "time"
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

//OsuBWTuple :- structure for bandwidth results
type OsuBWTuple struct {
    Bw      float64 `json:"bw"`
    Pktsize int     `json:"pktsize"`
}

//OsuBW :- List of bandwidth results
type OsuBW []OsuBWTuple

//OsuBiBW :- List of bidirectional bandwidth results
type OsuBiBW []OsuBWTuple

//OsuLatencyTuple :- Tuple for latency results
type OsuLatencyTuple struct {
    Latency float64 `json:"latency"`
    Pktsize int     `json:"pktsize"`
}

//OsuLatency :- Array of latency tuples
type OsuLatency []OsuLatencyTuple

//OSUResults :- OSU test result json structure.
// New test results should be added here to export
type OSUResults struct {
    Timestamp  time.Time `json:"timestamp"`
    OsuBW      `json:"OsuBW"`
    OsuBiBW    `json:"OsuBiBW"`
    OsuLatency `json:"OsuLatency"`
}

//Text2Json :- Structure + methods to generate matric
// results in json and matric file format
type Text2Json struct {
    jsonFile    string
    filelist    []string
    jsonResults *OSUResults
    configObj   *config.AppConfig
}

//GetAllFiles :- Function to collect all the OSU test result files.
// Mostly used by internal methods
func (txt2jsonObj *Text2Json) GetAllFiles(path string) error {
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

//IsLatencyFile :- Function to check osu result file contain
// latency data. We use the file name to identify latency results.
func (txt2jsonObj *Text2Json) IsLatencyFile(fileName string) bool {
    return strings.Contains(fileName, "osu_latency")
}

//IsBWFile :- Check if a OSU result file is a bandwidth file.
// Use file name to identify file contain bandwidth results.
func (txt2jsonObj *Text2Json) IsBWFile(fileName string) bool {
    return strings.Contains(fileName, "osu_bw")
}

//IsBiBWFile :- Check if OSU result file set have a bidirectional
// bandwidth test results.
func (txt2jsonObj *Text2Json) IsBiBWFile(fileName string) bool {
    return strings.Contains(fileName, "osu_bibw")
}

//ReadOSUBWFile :- Generic function to read the OSU BW result set.
//The function is being used to read any result files generated
// by OSU tests.
func (txt2jsonObj *Text2Json) ReadOSUBWFile(fileName string) (
    []OsuBWTuple, error) {
    logger := logging.GetLoggerInstance()
    fileName = strings.Trim(fileName, "\n")

    file, err := os.Open(fileName)
    if err != nil {
        logger.Error("Failed to open file %s", fileName)
        return nil, err
    }

    defer file.Close()
    bwresults := make([]OsuBWTuple, 0)
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
        var bwtuple OsuBWTuple
        bwtuple.Pktsize, _ = strconv.Atoi(lineArr[0])
        bwtuple.Bw, _ = strconv.ParseFloat(lineArr[1], 64)
        bwresults = append(bwresults, bwtuple)
    }
    return bwresults, errors.OP_SUCCESS
}

//ReadOSULatencyFile :- Generic function to read latency result set.
func (txt2jsonObj *Text2Json) ReadOSULatencyFile(fileName string,
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

// SetupApolloEnv :- Before creating any logs, its necessary to create all the apollo directory
// structure to report matric
func (txt2jsonObj *Text2Json) SetupApolloEnv() {
    var err error
    logger := logging.GetLoggerInstance()
    err = os.MkdirAll(config.DEFAULT_APOLLO_ENV_DIR, os.ModePerm)
    if err != nil {
        logger.Error("Failed to create/open apollo dir : %s, matric push may fail"+
            " err : %s", config.DEFAULT_APOLLO_ENV_DIR, err)
    }
}

//Init :- Must be called this function as constructor before using
// any functinalities of matric file generator.
func (txt2jsonObj *Text2Json) Init(configObj *config.AppConfig, resPath string) {
    txt2jsonObj.GetAllFiles(resPath)
    txt2jsonObj.jsonResults = new(OSUResults)
    txt2jsonObj.jsonFile = resPath + "osu-report.json"
    txt2jsonObj.configObj = configObj
    txt2jsonObj.SetupApolloEnv()
}

//Read2JsonStruct :- Reading the results to json format
func (txt2jsonObj *Text2Json) Read2JsonStruct() {
    var err error
    logger := logging.GetLoggerInstance()
    var latencyResults OsuLatency
    var bwresults []OsuBWTuple
    var bibwresults []OsuBWTuple
    for _, fileName := range txt2jsonObj.filelist {
        if txt2jsonObj.IsLatencyFile(fileName) {
            // Process only latency files
            txt2jsonObj.ReadOSULatencyFile(fileName, &latencyResults)
            txt2jsonObj.jsonResults.OsuLatency = latencyResults
            logger.Info("Processing of latency results  is complete")
        }
        if txt2jsonObj.IsBWFile(fileName) {
            //Process the bandwidth results
            bwresults, err = txt2jsonObj.ReadOSUBWFile(fileName)
            if err == errors.OP_SUCCESS {
                //Write data only when the bw result processing is success
                txt2jsonObj.jsonResults.OsuBW = bwresults
            }
        }
        if txt2jsonObj.IsBiBWFile(fileName) {
            bibwresults, err = txt2jsonObj.ReadOSUBWFile(fileName)
            if err == errors.OP_SUCCESS {
                //Write data only when the bw result processing is success
                txt2jsonObj.jsonResults.OsuBiBW = bibwresults
            }
        }
    }
}

//WriteTimestamp :- Write current timestamp to the output file
func (txt2jsonObj *Text2Json) WriteTimestamp() error {
    var err error
    fileSet := strings.Split(txt2jsonObj.jsonFile, "/")
    timestr := fileSet[len(fileSet)-3]
    txt2jsonObj.jsonResults.Timestamp, err = time.Parse(config.DEFAULT_TIME_LAYOUT,
        timestr)
    if err == nil {
        return errors.OP_SUCCESS
    }
    return err
}

//AppendBW2MatricOutput :- Function to populate matric data to a string.
// Parameters
// starttime :- time of the entry
// mclass    :- type of matric/entry, can be bandwidth or latency
// pktsize   :- size of packets used for measurements
// valuetype :- type of value, can be bidirectional/unidirectional
func (txt2jsonObj *Text2Json) AppendBW2MatricOutput(starttime time.Time,
    valuename string,
    bwvalues []OsuBWTuple,
    result *string) {
    *result = "--------------------------------------------\n" +
        "StartTime=" + strconv.FormatInt(starttime.Unix(), 10) + "\n" +
        "Host=" + txt2jsonObj.configObj.HostName + "\n" +
        "Marketplace=" + txt2jsonObj.configObj.Region + "\n" +
        "Program=ec2-osu-benchmark-bw\n" +
        "Time=0\n" +
        "Metrics="
    for _, entry := range bwvalues {
        *result = fmt.Sprintf("%s%s|%d|%s=%f,", *result,
            txt2jsonObj.configObj.HostName,
            entry.Pktsize,
            valuename, entry.Bw)
    }
    *result = *result + "\nEOE\n"
}

//GetMatricFileName : - Matric file name should have the specific timestamp information
// File get rolled in every hour based on the timestamp
func (txt2jsonObj *Text2Json) GetMatricFileName() string {
    t := time.Now()
    day := t.Format("2006-01-02")
    hour := t.Hour()
    return fmt.Sprintf("%s%s-%d", config.DEFAULT_MATRIC_OUTPUT_FILE_PREFIX,
        day, hour)

}

func (txt2jsonObj *Text2Json) _Write2MatricFile(result string) error {
    logger := logging.GetLoggerInstance()
    fileName := txt2jsonObj.GetMatricFileName()
    fp, err := os.OpenFile(fileName,
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0644)
    if err != nil {
        logger.Error("Failed to create/open matric file %s",
            fileName)
        return err
    }
    defer fp.Close()
    _, err = fp.Write([]byte(result))
    if err != nil {
        logger.Error("Failed to write results to file %s",
            fileName)
    }
    return errors.OP_SUCCESS
}

//Write2MatricFile :- Writing to matric file to export to the
// cloudwatch, the format is not same as json.
func (txt2jsonObj *Text2Json) Write2MatricFile() error {
    var bwresults string
    //var bibwresults string
    //var latencyresults string
    txt2jsonObj.AppendBW2MatricOutput(txt2jsonObj.jsonResults.Timestamp,
        "UniDirBWinMB",
        ([]OsuBWTuple)(txt2jsonObj.jsonResults.OsuBW),
        &bwresults)
    txt2jsonObj._Write2MatricFile(bwresults)
    return errors.OP_SUCCESS
}

//WriteJSONFile :- Function to write the structure to json result file.
func (txt2jsonObj *Text2Json) WriteJSONFile() error {
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

//ProcessResults2Json :- Process results to json format
func (txt2jsonObj *Text2Json) ProcessResults2Json() error {
    logger := logging.GetLoggerInstance()
    txt2jsonObj.WriteTimestamp()
    txt2jsonObj.Read2JsonStruct()
    err := txt2jsonObj.WriteJSONFile()
    if err != errors.OP_SUCCESS {
        logger.Error("Failed to write to json file")
    }
    err = txt2jsonObj.Write2MatricFile()
    return err
}
