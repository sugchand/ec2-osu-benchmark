package config

import (
    "os"
    "fmt"
    "flag"
    "ec2-osu-benchmark/logging"
    "ec2-osu-benchmark/errors"
)


type AppConfig struct {
    // Number of cores/processes to run the benchmark testing
    MPIcount uint
    //hostfile with all the hostnames involved in the testing
    HostFile string
    LogFile string 
    Loglevel int64
}

const (
    DEFAULT_LOG_LEVEL = logging.Trace
    DEFAULT_PATH = "/tmp/"
    DEFAULT_LOG_FILE = DEFAULT_PATH + "osu-test.log"
    DEFAULT_MPI_COUNT = 2
    DEFAULT_MPI_HOSTFILE = DEFAULT_PATH + "hostfile"
    DEFAULT_TIME_LAYOUT = "2006-01-02T15:04:05.999999-07:00"
)

func (config *AppConfig)printHelp() {
    helpstr := "\n\t OSU benchmark test running on EC2 instances" +
           "\n\t Running OSU MPI benchmark tests on EC2 instances " +
           "\n\t   USAGE: ./ec2-osu-benchmark {ARGS}" +
           "\n\t   ARGS:" +
           "\n\t    -help / -h                          :- Display help and exit." +
           "\n\t    -c <count> / -mpicount <count>      :- Number of MPI processes/cores" +
           "\n\t    -f <file> / -hostfile <file>        :- hostfile with MPI host info" +
           "\n\t    -l <loglevel>/ -loglevel <loglevel> :- loglevel for the application(Default :2)" +
           "\n\t                                             1. Trace" +
           "\n\t                                             2. Info" +
           "\n\t                                             3. Warning" +
           "\n\t                                             4. Error" +
           "\n\n"
    fmt.Print(helpstr)
}

//Read the config from the commandline to the config structure
func (config *AppConfig)InitConfig() error{
    var err error
    err = errors.OP_SUCCESS
    flag.Usage = config.printHelp
    mpicountShort := flag.Uint("c", DEFAULT_MPI_COUNT,
                              "Number of MPI processes/cores")
    mpicountLong := flag.Uint("mpicount", DEFAULT_MPI_COUNT,
                             "Number of MPI processes/cores")
    mpihostfileShort := flag.String("f", DEFAULT_MPI_HOSTFILE,
                                   "hostfile with MPI host info")
    mpihostfileLong := flag.String("hostfile", DEFAULT_MPI_HOSTFILE,
                                   "hostfile with MPI host info")
    loglevelShort := flag.Int64("l", DEFAULT_LOG_LEVEL,
                                "loglevel for the application")
    loglevellong := flag.Int64("loglevel", DEFAULT_LOG_LEVEL,
                               "loglevel for the application")
    flag.Parse()
    config.MPIcount = *mpicountShort
    if config.MPIcount == DEFAULT_MPI_COUNT {
        config.MPIcount = *mpicountLong
    }

    config.HostFile = *mpihostfileShort
    if config.HostFile == DEFAULT_MPI_HOSTFILE {
        config.HostFile = *mpihostfileLong
    }
    config.Loglevel = *loglevelShort
    if config.Loglevel == DEFAULT_LOG_LEVEL {
        config.Loglevel = *loglevellong
    }

    config.LogFile = DEFAULT_LOG_FILE
    // Check if hostfile exists in the filesystem
    if _, err := os.Stat(config.HostFile); os.IsNotExist(err) {
        fmt.Print("Hostfile is not present, cannot run tests \n")
        return err
    }
 
    fmt.Printf("\n*** Running test with cores/processes : %d , hostfile : %s " +
               "LogFile : %s, LogLevel %s ***\n",
               config.MPIcount, config.HostFile,
               config.LogFile, logging.LogLevelStr[config.Loglevel - 1])
    return err
}