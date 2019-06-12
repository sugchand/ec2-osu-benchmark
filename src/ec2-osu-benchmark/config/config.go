package config

import (
    "os"
    "fmt"
    "flag"
    "os/exec"
    "ec2-osu-benchmark/logging"
    "ec2-osu-benchmark/errors"
)


type AppConfig struct {
    HostName string
    // Number of cores/processes to run the benchmark testing
    MPIcount uint
    //hostfile with all the hostnames involved in the testing
    HostFile string
    Region string // Region at which the instance belongs to
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
    DEFAULT_REGION = "CMH52-CELL02340001"
    DEFAULT_APOLLO_ENV_DIR = "/apollo/env/OSU-MPI/monitoring/metricagent/"
    DEFAULT_MATRIC_OUTPUT_FILE_PREFIX = DEFAULT_APOLLO_ENV_DIR +
                                        "service_log."
)

func (config *AppConfig)printHelp() {
    helpstr := "\n\t OSU benchmark test running on EC2 instances" +
           "\n\t Running OSU MPI benchmark tests on EC2 instances " +
           "\n\t   USAGE: ./ec2-osu-benchmark {ARGS}" +
           "\n\t   ARGS:" +
           "\n\t    -help / -h                              :- Display help and exit." +
           "\n\t    -c <count> / -mpicount <count>          :- Number of MPI processes/cores" +
           "\n\t    -f <file> / -hostfile <file>            :- hostfile with MPI host info" +
           "\n\t    -r <region>/ -region <region>           :- Region of ec2 instance" +
           "\n\t    -l <loglevel>/ -loglevel <loglevel>     :- loglevel for the application(Default :2)" +
           "\n\t                                              1. Trace" +
           "\n\t                                              2. Info" +
           "\n\t                                              3. Warning" +
           "\n\t                                              4. Error" +
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
    regionShort := flag.String("r", DEFAULT_REGION,
                                "Region of ec2 instance")
    regionLong := flag.String("region", DEFAULT_REGION,
                                "Region of ec2 instance")
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

    config.Region = *regionShort
    if config.Region == DEFAULT_REGION {
        config.Region = *regionLong
    }

    // Populate the external host name of the instance
    var res []byte
    config.HostName="localhost"
    awsFindDNSCmd := "curl -s http://169.254.169.254/latest/meta-data/public-hostname"
    res, err = exec.Command("sh","-c", awsFindDNSCmd).Output()
    if err == nil {
        config.HostName = string(res)
    } else {
        fmt.Printf("Failed to collect hostname of ec-2 instance err : %s", err)
    }
    fmt.Printf("\n*** Running test on %s with cores/processes : %d , hostfile : %s, " +
               " Region %s, "+
               "LogFile : %s, LogLevel %s ***\n",
               config.HostName,
               config.MPIcount, config.HostFile,
               config.Region,
               config.LogFile, logging.LogLevelStr[config.Loglevel - 1])
    return errors.OP_SUCCESS
}