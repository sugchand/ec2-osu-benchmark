package testRunner

import (
    "fmt"
    "time"
    "strings"
    "os"
    "os/exec"
    "ec2-osu-benchmark/logging"
    "ec2-osu-benchmark/errors"
    "ec2-osu-benchmark/sys"
    "ec2-osu-benchmark/config"
)

//*****************************************************************************
// ************ NOTE ::: MAKE SURE COMMANDS ARE IN RIGHT PATH *****************
//*****************************************************************************
var osu_cmds = []string {
    "/usr/local/libexec/osu-micro-benchmarks/mpi/pt2pt/osu_latency",
    "/usr/local/libexec/osu-micro-benchmarks/mpi/pt2pt/osu_bw",
    "/usr/local/libexec/osu-micro-benchmarks/mpi/pt2pt/osu_bibw"}

// Expecting at max of 2000 results produced at a time.
var RESULT_CHANNEL_SIZE uint64 = 2000

type osu_result_channel struct {
    resultData string
    resultFileName string
}


type OSU_MPI_cmds struct {
    mpirunCmd string
    osu_cmds []string
    result_channel chan osu_result_channel
    result_channel_size uint64
    exit_result_write bool
    result_dir string
}

//*****************************************************************************
//********************  OSU command channel Functions *************************
//*****************************************************************************
func (chanObj *osu_result_channel)SetResultChannel(resultData string,
                                                   resultFileName string) {
    chanObj.resultData = resultData
    chanObj.resultFileName = resultFileName
}

func (chanObj *osu_result_channel)GetResultChannelData() (string, string) {
    resultData := chanObj.resultData
    resultFileName := chanObj.resultFileName
    return resultData, resultFileName
}

//*****************************************************************************
func (mpi_cmd_obj *OSU_MPI_cmds)IsCmdExists(execmd string) bool {
      cmd := exec.Command("/bin/sh", "-c", "command -v "+execmd)
      if err := cmd.Run(); err != nil {
              return false
      }
      return true
}

func (mpi_cmd_obj *OSU_MPI_cmds)Init_OSU_MPI_Cmds(MPIcount uint,
                                                  hostfile string) error {
    var err error
    err = errors.OP_SUCCESS
    logger := logging.GetLoggerInstance()
    if mpi_cmd_obj.IsCmdExists("mpirun") == false {
        logger.Error("Failed to find 'mpirun' in the system")
        return errors.CMD_NOT_FOUND
    }
    mpi_cmd_obj.mpirunCmd = fmt.Sprintf("mpirun --allow-run-as-root " +
                             "--np %d --hostfile %s",
                             MPIcount, hostfile)
    // TODO :: May need to allow selective command execution.
    mpi_cmd_obj.osu_cmds = osu_cmds
    mpi_cmd_obj.result_channel_size = RESULT_CHANNEL_SIZE
    mpi_cmd_obj.result_channel = make(chan osu_result_channel, 
                                        mpi_cmd_obj.result_channel_size)
    mpi_cmd_obj.exit_result_write = false
    result_dir := fmt.Sprintf("%s%s/%d/", config.DEFAULT_PATH, 
                            time.Now().Format(config.DEFAULT_TIME_LAYOUT),
                            os.Getpid())
    err = os.MkdirAll(result_dir, os.ModePerm)
    if err != nil {
        logger.Error("Failed to create result directory\n err : %s", err)
        return err
    }
    mpi_cmd_obj.result_dir = result_dir
    return errors.OP_SUCCESS
}

func (mpi_cmd_obj *OSU_MPI_cmds)get_cmd_fileName(cmd string) string{
    execCmd := strings.Split(cmd, "/")
    lastCmd := execCmd[len(execCmd) -1]
    lastCmd = mpi_cmd_obj.result_dir + lastCmd + ".txt"
    return lastCmd
}

func (mpi_cmd_obj *OSU_MPI_cmds)Run_OSU_MPI_Cmds() error {
    var err error
    var res []byte
    err = errors.OP_SUCCESS
    logger := logging.GetLoggerInstance()

    for _, cmd := range osu_cmds {
        if mpi_cmd_obj.IsCmdExists(cmd) == false {
            //Cannot find the command in the system.
            logger.Error("Failed to run command %s, as its not found", cmd)
            continue
        }
        run_cmd := fmt.Sprintf("%s %s", mpi_cmd_obj.mpirunCmd, cmd)
        logger.Info(" *** Running test command %s ***\n", run_cmd)
        res, err = exec.Command("sh","-c", run_cmd).Output()
        if err != nil {
            logger.Error("Failed to run test : %s, err : %s\n", run_cmd, err)
             // Continue with next test set
             continue
        }
        cmdFileName := mpi_cmd_obj.get_cmd_fileName(cmd)
        var res_channel osu_result_channel
        res_channel.SetResultChannel(string(res), cmdFileName)
        // Push to the channel for write go-routine
        mpi_cmd_obj.result_channel <- res_channel
    }
    return err
}

func (mpi_cmd_obj *OSU_MPI_cmds)write_to_file(result *osu_result_channel) error{
    var resultData, resultFileName string
    logger := logging.GetLoggerInstance()
    resultData, resultFileName = result.GetResultChannelData()
    fp, err := os.OpenFile(resultFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY,
                           0644)
    if err != nil {
        logger.Error("Failed to create result file %s", resultFileName)
        return err
    }
    defer fp.Close()
    _, err = fp.Write([]byte(resultData))
    if err != nil {
        logger.Error("Failed to write results to file %s", resultFileName)
    }

    return errors.OP_SUCCESS
}

func (mpi_cmd_obj *OSU_MPI_cmds)Get_OSU_MPI_test_result_path() string {
    return mpi_cmd_obj.result_dir
}

// Go routine to read command output and write to file stream.
func (mpi_cmd_obj *OSU_MPI_cmds)WriteCommandOutput() {
    for mpi_cmd_obj.exit_result_write == false {
        select {
            case osu_result := <- mpi_cmd_obj.result_channel:
                mpi_cmd_obj.write_to_file(&osu_result)
            default:
                // Do nothing
        }
    }
    // While exiting, make sure to mark the go-routine exit in sync
    syncObj := sys.GetAppSyncObj()
    syncObj.ExitRoutineInWaitGroup() 
}

func (mpi_cmd_obj *OSU_MPI_cmds)ExitresultWriteRoutine() {
    mpi_cmd_obj.exit_result_write = true
}
