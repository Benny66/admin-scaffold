package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构体
type Config struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name          string `yaml:"name"`            // 系统名称
	Subtitle      string `yaml:"subtitle"`        // 系统副标题
	Footer        string `yaml:"footer"`          // 页脚
	Logo          string `yaml:"logo"`            // 品牌 logo 文件名（放 backend/static/）
	Favicon       string `yaml:"favicon"`         // 浏览器标签图标文件名（放 backend/static/，缺省回退 logo）
	LoginBg       string `yaml:"login_bg"`        // 桌面端登录页背景图文件名（放 backend/static/，留空回退渐变）
	LoginBgMobile string `yaml:"login_bg_mobile"` // 移动端登录页背景图文件名（留空回退 login_bg）
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type   string       `yaml:"type"` // sqlite | mysql
	SQLite SQLiteConfig `yaml:"sqlite"`
	MySQL  MySQLConfig  `yaml:"mysql"`
}

// SQLiteConfig SQLite 配置
type SQLiteConfig struct {
	DSN string `yaml:"dsn"`
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
}

// DSN 生成 MySQL DSN 连接字符串
func (m MySQLConfig) DSN() string {
	charset := m.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		m.Username, m.Password, m.Host, m.Port, m.Database, charset)
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret        string `yaml:"secret"`
	ExpireSeconds int    `yaml:"expire_seconds"`
}

// GlobalConfig 全局配置实例
var GlobalConfig Config

func init() {
	// 先设置默认值
	GlobalConfig = Config{
		App: AppConfig{
			Name:          "",
			Subtitle:      "",
			Footer:        "",
			Logo:          "logo.png",
			Favicon:       "",
			LoginBg:       "",
			LoginBgMobile: "",
		},
		Server: ServerConfig{
			Port: "8080",
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			SQLite: SQLiteConfig{
				DSN: "./base_backend.db",
			},
			MySQL: MySQLConfig{
				Host:     "127.0.0.1",
				Port:     "3306",
				Username: "root",
				Password: "",
				Database: "base_backend",
				Charset:  "utf8mb4",
			},
		},
		JWT: JWTConfig{
			Secret:        "base-backend-secret-key-change-me",
			ExpireSeconds: 3600,
		},
	}

	// 尝试从 YAML 配置文件加载
	loadYAMLConfig()

	// 环境变量覆盖（向后兼容）
	if v := os.Getenv("BASE_BACKEND_SERVER_PORT"); v != "" {
		GlobalConfig.Server.Port = v
	}
	if v := os.Getenv("BASE_BACKEND_GIN_MODE"); v != "" {
		GlobalConfig.Server.Mode = v
	}
	if v := os.Getenv("BASE_BACKEND_DB_TYPE"); v != "" {
		GlobalConfig.Database.Type = v
	}
	if v := os.Getenv("BASE_BACKEND_DB_DSN"); v != "" {
		GlobalConfig.Database.Type = "sqlite"
		GlobalConfig.Database.SQLite.DSN = v
	}
	if v := os.Getenv("BASE_BACKEND_JWT_SECRET"); v != "" {
		GlobalConfig.JWT.Secret = v
	}
	if v := os.Getenv("BASE_BACKEND_JWT_EXPIRE_SECONDS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			GlobalConfig.JWT.ExpireSeconds = i
		}
	}

	// SQLite 相对路径解析（基于可执行文件目录）
	if GlobalConfig.Database.Type == "sqlite" {
		GlobalConfig.Database.SQLite.DSN = resolveDSN(GlobalConfig.Database.SQLite.DSN)
	}

	// 启动时打印实际生效的 JWT 配置，方便排查
	secretMask := GlobalConfig.JWT.Secret
	if len(secretMask) > 4 {
		secretMask = secretMask[:4] + "***"
	}
	log.Printf("[启动] JWT 配置: expire_seconds=%d, secret=%s", GlobalConfig.JWT.ExpireSeconds, secretMask)
}

// yamlFile YAML 配置文件结构（用于反序列化）
type yamlFile struct {
	App struct {
		Name          string `yaml:"name"`
		Subtitle      string `yaml:"subtitle"`
		Footer        string `yaml:"footer"`
		Logo          string `yaml:"logo"`
		Favicon       string `yaml:"favicon"`
		LoginBg       string `yaml:"login_bg"`
		LoginBgMobile string `yaml:"login_bg_mobile"`
	} `yaml:"app"`
	Server struct {
		Port string `yaml:"port"`
		Mode string `yaml:"mode"`
	} `yaml:"server"`
	Database struct {
		Type   string `yaml:"type"`
		SQLite struct {
			DSN string `yaml:"dsn"`
		} `yaml:"sqlite"`
		MySQL struct {
			Host     string `yaml:"host"`
			Port     string `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			Database string `yaml:"database"`
			Charset  string `yaml:"charset"`
		} `yaml:"mysql"`
	} `yaml:"database"`
	JWT struct {
		Secret        string `yaml:"secret"`
		ExpireSeconds int    `yaml:"expire_seconds"`
	} `yaml:"jwt"`
}

// loadYAMLConfig 从 config.yaml 加载配置
func loadYAMLConfig() {
	// 优先使用可执行文件同目录的 config.yaml
	configPath := "config.yaml"
	execPath, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(execPath), "config.yaml")
		if _, err := os.Stat(candidate); err == nil {
			configPath = candidate
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("未找到配置文件 %s，使用默认配置", configPath)
		return
	}

	var yf yamlFile
	if err := yaml.Unmarshal(data, &yf); err != nil {
		log.Printf("解析配置文件失败: %v，使用默认配置", err)
		return
	}

	// 覆盖默认值（仅覆盖非空字段）
	if yf.App.Name != "" {
		GlobalConfig.App.Name = yf.App.Name
	}
	if yf.App.Subtitle != "" {
		GlobalConfig.App.Subtitle = yf.App.Subtitle
	}
	if yf.App.Footer != "" {
		GlobalConfig.App.Footer = yf.App.Footer
	}
	if yf.App.Logo != "" {
		GlobalConfig.App.Logo = yf.App.Logo
	}
	if yf.App.Favicon != "" {
		GlobalConfig.App.Favicon = yf.App.Favicon
	}
	if yf.App.LoginBg != "" {
		GlobalConfig.App.LoginBg = yf.App.LoginBg
	}
	if yf.App.LoginBgMobile != "" {
		GlobalConfig.App.LoginBgMobile = yf.App.LoginBgMobile
	}
	if yf.Server.Port != "" {
		GlobalConfig.Server.Port = yf.Server.Port
	}
	if yf.Server.Mode != "" {
		GlobalConfig.Server.Mode = yf.Server.Mode
	}
	if yf.Database.Type != "" {
		GlobalConfig.Database.Type = yf.Database.Type
	}
	if yf.Database.SQLite.DSN != "" {
		GlobalConfig.Database.SQLite.DSN = yf.Database.SQLite.DSN
	}
	if yf.Database.MySQL.Host != "" {
		GlobalConfig.Database.MySQL.Host = yf.Database.MySQL.Host
	}
	if yf.Database.MySQL.Port != "" {
		GlobalConfig.Database.MySQL.Port = yf.Database.MySQL.Port
	}
	if yf.Database.MySQL.Username != "" {
		GlobalConfig.Database.MySQL.Username = yf.Database.MySQL.Username
	}
	if yf.Database.MySQL.Password != "" {
		GlobalConfig.Database.MySQL.Password = yf.Database.MySQL.Password
	}
	if yf.Database.MySQL.Database != "" {
		GlobalConfig.Database.MySQL.Database = yf.Database.MySQL.Database
	}
	if yf.Database.MySQL.Charset != "" {
		GlobalConfig.Database.MySQL.Charset = yf.Database.MySQL.Charset
	}
	if yf.JWT.Secret != "" {
		GlobalConfig.JWT.Secret = yf.JWT.Secret
	}
	if yf.JWT.ExpireSeconds > 0 {
		GlobalConfig.JWT.ExpireSeconds = yf.JWT.ExpireSeconds
	}

	log.Printf("已加载配置文件: %s", configPath)
}

// resolveDSN 将相对路径的 DSN 转为相对于可执行文件所在目录的绝对路径
func resolveDSN(dsn string) string {
	if filepath.IsAbs(dsn) {
		return dsn
	}
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("警告: 无法获取可执行文件路径，使用原始DSN: %v", err)
		return dsn
	}
	execDir := filepath.Dir(execPath)
	// go run 时 os.Executable() 返回临时目录，改用当前工作目录
	if strings.Contains(execDir, "go-build") || strings.HasPrefix(execDir, os.TempDir()) {
		cwd, err := os.Getwd()
		if err == nil {
			log.Printf("检测到 go run 模式，使用工作目录: %s", cwd)
			return filepath.Join(cwd, dsn)
		}
	}
	resolved := filepath.Join(execDir, dsn)
	log.Printf("数据库路径: %s", resolved)
	return resolved
}
