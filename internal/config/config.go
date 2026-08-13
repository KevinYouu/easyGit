package config

type Option struct {
	Label       string
	Value       string
	Usage       int
	Description string // 选项单行说明(为空则渲染纯名称)
}

type Patch struct {
	Prefix string
	Major  int
	Minor  int
	Patch  int
	Suffix string
}
