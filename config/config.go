package config

import (
	"fmt"
	"time"
)

type Config struct {
	MySQL MySQLConfig
	App   AppConfig
}

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type AppConfig struct {
	AntpoolCookie  string
	RequestTimeout time.Duration
	MinerUser      string
	MinerPassword  string
}

func Load() *Config {
	// In a real application, you would load this from env or a file (viper)
	// Based on requirements:
	return &Config{
		MySQL: MySQLConfig{
			Host:     "8.137.94.107",
			Port:     "3306", // Default port, not specified but assumed
			User:     "admin",
			Password: "admin$2026",
			Database: "myapp_db",
		},
		App: AppConfig{
			// Cookie from requirements
			AntpoolCookie:  `_uab_collina=176543110684584189051945; _ga=GA1.1.2019618821.1765431115; _ga_0ZDBQJ5SB7=GS2.1.s1766753979$o1$g1$t1766754049$j58$l0$h0; tfstk=gbu-QC9_8KvoSEYXBaxDKeeE2jADincyZYl1-J2lAxHx1fWuA3l3OXMj3yqBUz0KHYk4q00KTwFIOvRzKQ-mabzURdviJFcraP_fAlp0RBGb-WLET84SabzFgxHq5WGyJD397WaIdrZb_WzCV7wWMoNgOy_7AaGbh-PQRyZ5Powb95S5RgsWMjw4OywIdyOYl-PQRJMQRApfwJf7pwnuCu3CkpAm-2gYwuesMn7CJY_g2RiYHw9sk7BQCbwARww_SeRoMbvR7j4rSAFiEE_-BXGZAo3dhdeEkjgSAjWJdSoIErwu531Yr4lSfkg6LMMTyvEsPoCCyraYdrFj5Bjanqeod4EXsNE3lVq_Pmx2Ek486vgre6QICjcizoupBtwEq5zb92R1vJEC4D0iW_WfIRFhVIdAYMr7g71hLO24UT0UMRAcmMSUmSPYIIdAYMr7gSeMiEjFYoVV.; _c_WBKFRo=w3XSPKSsoD4FjHfWTDyzqITMtrRfNAIcQRD5kE5y; language=zh; informedConsent=1; wwwroute=1769654858.882.1806.809786; JSESSIONID=AF3D4670959027C5F83143C3B67739B0; acw_tc=ac11000117696748984088047e2d4bbf77c97a5515fee3b3a48eb31351eae5; _ga_Y4GT96XRF0=GS2.1.s1769675628$o62$g0$t1769675628$j60$l0$h0`,
			RequestTimeout: 30 * time.Second,
			MinerUser:      "root",
			MinerPassword:  "root",
		},
	}
}

func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database)
}
