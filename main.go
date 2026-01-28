package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type Rule struct {
	Method string `json:"method"`
	Value  string `json:"value"`
}

type RuleSet struct {
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}

type Config struct {
	RuleSets []RuleSet `json:"-"`
}

type MatchResult struct {
	RuleName string
	Rule     Rule
	Input    string
	Matched  bool
	Error    error
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var ruleSets []RuleSet
	if err := json.Unmarshal(data, &ruleSets); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &Config{RuleSets: ruleSets}, nil
}

func (c *Config) TestRegex(input string) []MatchResult {
	var results []MatchResult

	for _, ruleSet := range c.RuleSets {
		for _, rule := range ruleSet.Rules {
			result := MatchResult{
				RuleName: ruleSet.Name,
				Rule:     rule,
				Input:    input,
			}

			switch rule.Method {
			case "regex":
				re, err := regexp.Compile(rule.Value)
				if err != nil {
					result.Error = fmt.Errorf("编译正则表达式失败: %w", err)
				} else {
					result.Matched = re.MatchString(input)
				}
			case "not_regex":
				re, err := regexp.Compile(rule.Value)
				if err != nil {
					result.Error = fmt.Errorf("编译正则表达式失败: %w", err)
				} else {
					result.Matched = !re.MatchString(input)
				}
			case "contains":
				result.Matched = strings.Contains(input, rule.Value)
			case "not_contains":
				result.Matched = !strings.Contains(input, rule.Value)
			default:
				result.Error = fmt.Errorf("不支持的匹配方法: %s", rule.Method)
			}

			results = append(results, result)
		}
	}

	return results
}

func (c *Config) TestAllRules(input string) (bool, []MatchResult) {
	results := c.TestRegex(input)

	allMatched := true
	hasError := false

	for _, result := range results {
		if result.Error != nil {
			hasError = true
			break
		}
		if !result.Matched {
			allMatched = false
		}
	}

	if hasError {
		return false, results
	}

	return allMatched, results
}

func (c *Config) PrintRules() {
	fmt.Println("=== 配置文件中的规则 ===")
	for _, ruleSet := range c.RuleSets {
		fmt.Printf("规则集: %s\n", ruleSet.Name)
		for i, rule := range ruleSet.Rules {
			fmt.Printf("  %d. 方法: %s, 模式: %s\n", i+1, rule.Method, rule.Value)
		}
		fmt.Println()
	}
}

type TestCase struct {
	URL    string
	Expect string
}

func LoadTestCases(filename string) ([]TestCase, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取测试用例文件失败: %w", err)
	}

	var testCases []TestCase
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 按制表符或空格分割
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			testCases = append(testCases, TestCase{
				URL:    fields[0],
				Expect: fields[1],
			})
		}
	}

	return testCases, nil
}

func printUsage() {
	fmt.Println("用法: go run main.go [选项]")
	fmt.Println("选项:")
	fmt.Println("  -c string    配置文件路径 (默认: config.json)")
	fmt.Println("  -t string    测试用例文件路径 (默认: test_case.txt)")
	fmt.Println("  -u string    单个URL进行测试")
	fmt.Println("  -h           显示帮助信息")
	fmt.Println("  -v           显示详细匹配结果")
	fmt.Println("")
	fmt.Println("示例:")
	fmt.Println("  go run main.go -c myconfig.json -t mytest.txt")
	fmt.Println("  go run main.go -u \"/api/v1/users?access_token=abc123\"")
	fmt.Println("  go run main.go -c rules.json -u \"/rest/products\"")
}

func testSingleURL(config *Config, url string, verbose bool) {
	fmt.Printf("测试单个URL: %s\n", url)
	allMatched, results := config.TestAllRules(url)

	// 计算实际结果
	actualStatus := "不匹配"
	if allMatched {
		actualStatus = "匹配"
	}

	fmt.Printf("匹配结果: %s\n", actualStatus)

	// 显示每个规则的详细结果
	if verbose {
		fmt.Println("详细匹配结果:")
		for _, result := range results {
			status := "未匹配"
			if result.Matched {
				status = "匹配"
			}

			if result.Error != nil {
				fmt.Printf("  %s - %s: 错误 - %v\n", result.RuleName, result.Rule.Method, result.Error)
			} else {
				fmt.Printf("  %s - %s: %s\n", result.RuleName, result.Rule.Method, status)
			}
		}
	}
}

func testFromFile(config *Config, testFile string, verbose bool) {
	// 从文件加载测试用例
	testCases, err := LoadTestCases(testFile)
	if err != nil {
		log.Fatalf("加载测试用例失败: %v", err)
	}

	fmt.Printf("=== 加载了 %d 个测试用例 ===\n", len(testCases))

	fmt.Println("=== 测试结果 ===")
	passed := 0
	failed := 0

	for _, testCase := range testCases {
		fmt.Printf("\n测试输入: %s\n", testCase.URL)
		if verbose {
			fmt.Printf("期望结果: %s\n", testCase.Expect)
		}

		allMatched, results := config.TestAllRules(testCase.URL)

		// 计算实际结果
		actualStatus := "不匹配"
		if allMatched {
			actualStatus = "匹配"
		}
		if verbose {
			fmt.Printf("实际结果: %s\n", actualStatus)
		}

		// 验证结果
		if actualStatus == testCase.Expect {
			fmt.Printf("✓ 测试通过\n")
			passed++
		} else {
			fmt.Printf("✗ 测试失败\n")
			failed++
		}

		// 显示每个规则的详细结果
		if verbose {
			fmt.Println("详细匹配结果:")
			for _, result := range results {
				status := "未匹配"
				if result.Matched {
					status = "匹配"
				}

				if result.Error != nil {
					fmt.Printf("  %s - %s: 错误 - %v\n", result.RuleName, result.Rule.Method, result.Error)
				} else {
					fmt.Printf("  %s - %s: %s\n", result.RuleName, result.Rule.Method, status)
				}
			}
		}
	}

	fmt.Printf("\n=== 测试统计 ===\n")
	fmt.Printf("总测试用例: %d\n", len(testCases))
	fmt.Printf("通过: %d\n", passed)
	fmt.Printf("失败: %d\n", failed)
	fmt.Printf("通过率: %.1f%%\n", float64(passed)/float64(len(testCases))*100)

	if failed == 0 {
		fmt.Println("✓ 所有测试用例通过!")
	} else {
		fmt.Printf("✗ 有 %d 个测试用例失败\n", failed)
	}
}

func main() {
	// 定义命令行参数
	configFile := flag.String("c", "config.json", "配置文件路径")
	testFile := flag.String("t", "test_case.txt", "测试用例文件路径")
	singleURL := flag.String("u", "", "单个URL进行测试")
	help := flag.Bool("h", false, "显示帮助信息")
	verbose := flag.Bool("v", false, "显示详细匹配结果")

	flag.Parse()

	// 显示帮助信息
	if *help {
		printUsage()
		return
	}

	// 加载配置
	config, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	config.PrintRules()

	// 根据参数选择测试模式
	if *singleURL != "" {
		// 单个URL测试模式
		testSingleURL(config, *singleURL, *verbose)
	} else {
		// 文件测试模式
		testFromFile(config, *testFile, *verbose)
	}
}
