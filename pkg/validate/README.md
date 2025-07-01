# 结构体验证器

这个包提供了基于`go-playground/validator/v10`的功能强大的验证工具，支持结构体、单个变量、映射和切片的验证，并提供详细的错误信息和中文翻译。

## 功能特点

- 支持结构体标签验证
- 支持单个变量验证
- 支持映射(Map)验证
- 支持切片(Slice)验证
- 提供详细的验证错误信息
- 支持中文错误消息
- 支持自定义验证规则
- 简单易用的API

## 安装

```bash
go get github.com/go-playground/validator/v10
```

## 基本使用

### 1. 结构体验证

结构体验证是最常见的用法，通过在结构体字段上添加`validate`标签来定义验证规则：

```go
type User struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"required,gte=18,lte=120"`
	Password string `json:"password" validate:"required,min=8"`
}

// 验证结构体
user := &User{...}
if err := validator.ValidateStruct(user); err != nil {
	fmt.Printf("验证失败: %v\n", err)
}
```

### 2. 单个变量验证

单个变量验证可以对任何类型的单个变量进行验证：

```go
// 验证电子邮件
email := "test@example.com"
if err := validator.ValidateVar(email, "email"); err != nil {
	fmt.Printf("无效的电子邮件: %v\n", err)
}

// 验证数字范围
age := 15
if err := validator.ValidateVar(age, "gte=18,lte=120"); err != nil {
	fmt.Printf("年龄不在有效范围内: %v\n", err)
}

// 验证变量之间的关系
min := 10
max := 5
if err := validator.ValidateVarWithValue(max, min, "gtefield"); err != nil {
	fmt.Printf("最大值必须大于等于最小值: %v\n", err)
}
```

### 3. 映射(Map)验证

映射验证可以验证映射中的每个值（如果值是结构体）：

```go
userMap := map[string]interface{}{
	"user1": &User{...}, // 有效用户
	"user2": &User{...}, // 无效用户
}

if err := validator.ValidateMap(userMap); err != nil {
	fmt.Printf("映射验证失败: %v\n", err)
}
```

### 4. 切片(Slice)验证

切片验证可以验证切片中的每个元素（如果元素是结构体）：

```go
users := []*User{
	&User{...}, // 有效用户
	&User{...}, // 无效用户
}

if err := validator.ValidateSlice(users); err != nil {
	fmt.Printf("切片验证失败: %v\n", err)
}
```

### 5. 获取详细错误信息

如果需要获取更详细的错误信息，可以将错误转换为`ValidationErrors`类型：

```go
if err := validator.ValidateStruct(user); err != nil {
	if valErrs, ok := err.(validator.ValidationErrors); ok {
		// 遍历所有错误
		for _, e := range valErrs {
			fmt.Printf("字段: %s, 标签: %s, 值: %v, 错误: %s\n", 
				e.Field, e.Tag, e.Value, e.Message)
		}
		
		// 获取字段错误映射
		fieldErrors := validator.GetFieldErrors(valErrs)
		for field, msg := range fieldErrors {
			fmt.Printf("%s: %s\n", field, msg)
		}
	}
}
```

### 6. 使用验证器实例

如果需要更多控制，可以创建一个验证器实例：

```go
// 创建验证器
v := validator.New()

// 验证结构体
errors, err := v.Validate(user)
if err != nil {
	fmt.Printf("验证过程中发生错误: %v\n", err)
} else if len(errors) > 0 {
	for _, e := range errors {
		fmt.Printf("字段: %s, 错误: %s\n", e.Field, e.Message)
	}
}
```

### 7. 自定义验证规则

你可以注册自定义验证规则：

```go
v := validator.New()

// 注册自定义验证规则
v.RegisterCustomValidation("chinese_mobile", func(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// 简单的中国手机号验证：以1开头的11位数字
	matched, _ := regexp.MatchString(`^1\d{10}$`, value)
	return matched
}, "{0}必须是有效的中国手机号码")

// 使用自定义验证规则
type Contact struct {
	Mobile string `validate:"required,chinese_mobile"`
}
```

## 常用验证标签

- `required`: 必填字段
- `email`: 电子邮件格式
- `min=n`: 最小长度/值
- `max=n`: 最大长度/值
- `gte=n`: 大于等于
- `lte=n`: 小于等于
- `oneof=a b c`: 枚举值，必须是列出的值之一
- `len=n`: 固定长度
- `numeric`: 数字
- `alpha`: 字母
- `alphanum`: 字母和数字
- `omitempty`: 如果字段为空，则跳过其他验证
- `eqfield=Field`: 必须等于指定字段的值
- `gtfield=Field`: 必须大于指定字段的值
- `ltfield=Field`: 必须小于指定字段的值

更多标签请参考 [validator文档](https://pkg.go.dev/github.com/go-playground/validator/v10)。

## 完整示例

完整示例请参考：`example/validator/simple/simple_example.go`。 