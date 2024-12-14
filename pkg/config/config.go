package config

type Config struct {
	Database Database
}

type Database struct {
	User        string
	Password    string
	Host        string
	Port        int
	Database    string
	Connections int32
}
