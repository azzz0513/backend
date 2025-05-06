package settings

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Conf 全局变量，用来保存程序所有的配置信息
var Conf = new(Config)

type Config struct {
	*AppConfig   `mapstructure:"app"`
	*LogConfig   `mapstructure:"log"`
	*MySQLConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
}

type AppConfig struct {
	Name      string `mapstructure:"name"`
	Mode      string `mapstructure:"mode"`
	Version   string `mapstructure:"version"`
	StartTime string `mapstructure:"start_time"`
	Port      int    `mapstructure:"port"`
	MachineID int64  `mapstructure:"machine_id"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DbName       string `mapstructure:"dbname"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func Init() (err error) {
	// 读取配置文件
	viper.SetConfigFile("./conf/config.yaml")
	//viper.SetConfigName("config") // 指定配置文件名称（无扩展名）
	//viper.SetConfigType("yaml")   // 如果配置文件的名称中没有扩展名，则需要配置此项（专用于从远程获取配置信息时指定配置文件）
	//viper.AddConfigPath("./conf") // 指定查找配置文件所在的路径
	//viper.SetConfigFile(filename) // 通过命令行输入配置文件路径

	// 查找并读取配置文件
	if err = viper.ReadInConfig(); err != nil {
		// 处理读取配置文件的错误
		fmt.Printf("viper.ReadInconfig() failed, err: %s\n", err)
		return
	}
	fmt.Println(viper.AllSettings())
	// 把读取到的序列信息反序列化到Conf变量中
	if err = viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal() failed, err: %s\n", err)
	}
	// 实时监控配置文件的变化
	viper.WatchConfig()
	// 当配置文件变化后调用一个回调函数
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 配置文件修改后程序需要做的业务
		fmt.Println("配置文件修改了...")
		if err := viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal() failed, err: %s\n", err)
		}
	})
	return
}
