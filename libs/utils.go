package libs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	md5simd "github.com/minio/md5-simd"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

func RunForever() error {
	var err error
	signals := []os.Signal{os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT,
		syscall.SIGUSR2, syscall.SIGKILL}
	sCtx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(sCtx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals...)
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case sig := <-signalChan:
			fmt.Println("receive sig: ", sig)
			switch sig {
			case syscall.SIGUSR2:
				// TODO
				return nil
			default:
				return nil
			}
		}
	})
	if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		cancel()
		return err
	}
	cancel()
	return nil
}

func RandStr(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range b {
		result[i] = charSet[int(b[i])%len(charSet)]
	}
	return string(result)
}

func Int64TStr(n int64) string {
	return strconv.FormatInt(n, 10)
}

func Float64TStr(n float64, precision int) string {
	return strconv.FormatFloat(n, 'f', precision, 64)
}

func StrTInt(s string) int {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		fmt.Println("Error converting string to int32: ", s, err)
		return 0
	}
	return int(i)
}

func StrTFloat64(s string) float64 {

	floatValue, err := strconv.ParseFloat(s, 64)
	if err != nil {
		fmt.Println("error converting string to float64: ", s, err)
		return 0.0
	}
	return floatValue
}

func Int32Sort(arr []int32) []int32 {
	intArr := make([]int, len(arr))
	for i, v := range arr {
		intArr[i] = int(v)
	}

	// 排序
	sort.Ints(intArr)

	// 将排序后的 int 切片转换回 int32 切片
	for i, v := range intArr {
		arr[i] = int32(v)
	}
	return arr
}

func Hash(text string) string {

	server := md5simd.NewServer()
	md5Hash := server.NewHash()
	_, _ = md5Hash.Write([]byte(text))
	digest := md5Hash.Sum([]byte{})
	encrypted := hex.EncodeToString(digest)

	server.Close()
	md5Hash.Close()

	return encrypted
}

func Sign(apiKey string, params map[string][]string) string {
	sortedParams := make([]string, 0, len(params))
	for key := range params {
		sortedParams = append(sortedParams, key)
	}
	sort.Strings(sortedParams)

	var paramStr string
	for _, key := range sortedParams {
		value := params[key][0]
		if value != "" {
			if paramStr != "" {
				paramStr += "&"
			}
			paramStr += key + "=" + value
		}
	}

	signStr := apiKey + "&" + paramStr
	sign := Hash(signStr)
	//fmt.Println("Sign string:", signStr)
	//fmt.Println("Sign:", sign)
	return sign
}

// MaskString 对字符串(user_id / user_name / 昵称等)做统一脱敏，输出恒为 10 位。
// 规则：
//   - 长度 > 4：前面固定 6 个 *，保留末尾 4 位，例 agent001winner333 -> ******r333
//   - 长度 <= 4：前面补 * 到 10 位，原串放尾部，例 abcd -> ******abcd，ab -> ********ab
//   - 空串直接返回空串
func MaskString(s string) string {
	runes := []rune(s)
	length := len(runes)
	if length == 0 {
		return ""
	}
	const total = 10
	if length > 4 {
		return strings.Repeat("*", 6) + string(runes[length-4:])
	}
	return strings.Repeat("*", total-length) + s
}
