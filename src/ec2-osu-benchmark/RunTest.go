package main

import (
    "fmt"
    "ec2-osu-benchmark/config"
    "ec2-osu-benchmark/logging"
    "ec2-osu-benchmark/testRunner"
    "ec2-osu-benchmark/sys"
    "ec2-osu-benchmark/errors"
    "ec2-osu-benchmark/text2json"
)

func startLoggerService(configObj *config.AppConfig) {
    logger := new(logging.Logging)
    logger.LogInitSingleton(logging.LogLeveltype(configObj.Loglevel),
                            configObj.LogFile)
    logger.Trace("Logging service is started..")
}

func Runtests(configObj *config.AppConfig,
              osu_mpi_tests *testRunner.OSU_MPI_cmds) error{
    
    err := osu_mpi_tests.Init_OSU_MPI_Cmds(configObj.MPIcount,
                                    configObj.HostFile)
    if err != errors.OP_SUCCESS {
        fmt.Print(err)
        return err
    }
    syncObj := sys.GetAppSyncObj()
    syncObj.AddRoutineInWaitGroup() 
    //Start the result writer thread
    go osu_mpi_tests.WriteCommandOutput()

    // Run the OSU test cases
    osu_mpi_tests.Run_OSU_MPI_Cmds()
    return errors.OP_SUCCESS
}
func Write2Json(path string) {
    jsonwrite := new(text2json.Text2Json)
    jsonwrite.Init(path)
    jsonwrite.ProcessResults2Json()
}

func main() {
    var err error
    configObj := new(config.AppConfig)
    err = configObj.InitConfig()
    if err !=  errors.OP_SUCCESS {
        // Failed to initialize configuration.
        fmt.Print("Failed to init, wrong configuration, exiting..\n")
        panic ("Exiting the testrun due to invalid configuration")
    }
    startLoggerService(configObj)
    syncObj := sys.GetAppSyncObj()
    osu_mpi_tests := new(testRunner.OSU_MPI_cmds)
    err = Runtests(configObj, osu_mpi_tests)
    if err != errors.OP_SUCCESS {
        panic ("Exiting the testrun due to failed to run tests")
    }
    osu_mpi_tests.ExitresultWriteRoutine()
    syncObj.JoinAllRoutines()
    // Write to json only after all go-routines are done with its processing
    Write2Json(osu_mpi_tests.Get_OSU_MPI_test_result_path())
}