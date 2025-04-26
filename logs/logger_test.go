package logs

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// 开始测试
	fmt.Println("开始测试...")
	// 在此处可以添加测试前的初始化代码

	_ = NewLogManager(nil, "dev")

	exitCode := m.Run()

	// 清理操作
	fmt.Println("测试清理...")

	// 退出测试
	fmt.Println("退出测试...")
	os.Exit(exitCode)
}

func TestLogger_tele(t *testing.T) {
	logger := GetLogger("gameHub.tele")

	logger.LogT("telegram test for info")
	logger.ErrorT("telegram test for info")
	logger.WarnT("telegram test for info")
}

func TestLogger_common(t *testing.T) {
	logger := GetLogger("common")

	logger.Log("test for info")
	logger.Error("test for info")
	logger.Warn("test for info")
	logger.Info("test for info")
	logger.Debug("test for info")
}

func TestLogger_dot(t *testing.T) {
	logger := GetLogger("gameHub.ops")

	logger.Log("test for info")
	logger.Error("test for info")
	logger.Warn("test for info")
	logger.Info("test for info")
	logger.Debug("test for info")
}

func TestLogger_Level_log(t *testing.T) {
	logger := GetLogger("gameHub.log")
	logger.SetLevel(LogLevel)

	logger.Log("test for info")
	logger.Error("test for info")
	logger.Warn("test for info")
	logger.Info("test for info")
	logger.Debug("test for info")
}

func TestLogger_Level_Error(t *testing.T) {
	logger := GetLogger("gameHub.error")
	logger.SetLevel(ErrorLevel)

	logger.Log("test for info")
	logger.Error("test for info")
	logger.Warn("test for info")
	logger.Info("test for info")
	logger.Debug("test for info")
}

func TestLogger_Level_Warn(t *testing.T) {
	logger := GetLogger("gameHub.warn")
	logger.SetLevel(WarnLevel)

	logger.Log("test for info")
	logger.Error("test for info")
	logger.Warn("test for info")
	logger.Info("test for info")
	logger.Debug("test for info")
}

func TestLogger_Level_Info(t *testing.T) {
	logger := GetLogger("gameHub.info")
	logger.SetLevel(InfoLevel)

	logger.Log("test for info")
	logger.Error("test for info")
	logger.Warn("test for info")
	logger.Info("test for info")
	logger.Debug("test for info")
}

func TestLogger_Level_Debug(t *testing.T) {
	logger := GetLogger("gameHub.debug")
	logger.SetLevel(DebugLevel)

	logger.Log("test for info")
	logger.Error("test for info")
	logger.Warn("test for info")
	logger.Info("test for info")
	logger.Debug("test for info debug....")
}

func TestLogger_Level_Off(t *testing.T) {
	logger := GetLogger("gameHub.log.off")
	logger.SetLevel(OffLevel)

	logger.Log("test for info")
	logger.Error("test for info")
	logger.Warn("test for info")
	logger.Info("test for info")
	logger.Debug("test for info")
}
