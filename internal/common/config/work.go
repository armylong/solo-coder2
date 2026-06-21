package config

// 飞书多维表格配置
const (
	FeishuDocAppToken = `CE3BwYISBiEG4KkG04UcTfr6nRh` // 多维表格AppToken
	FeishuDocTableId  = `tbliWHNKeW9dcQnw`            // 多维表格TableId
	FeishuDocViewId   = `vewGNON9rb`                  // 多维表格ViewId
)

// 工作目录配置
const (
	WorkHome           = `/root/works/doubao_testing`
	WorkSpace          = WorkHome + `/works`
	WorkFileName       = `work.json`   // 工作数据文件名
	WorkDoneFileName   = `work.done`   // 工作完成标记文件名
	QaDoneFileName     = `qa.done`     // 质检完成标记文件名
	QaFileName         = `qa.json`     // 质检数据文件名
	UploadDoneFileName = `upload.done` // 上传完成标记文件名
)

// 条件运算符-整体
var (
	ConjunctionAnd string = `and` // 且
)

// 条件运算符-单条
var (
	OperatorIs             string = `is`             // 等于
	OperatorNot            string = `isNot`          // 不等于
	OperatorContains       string = `contains`       // 包含
	OperatorDoesNotContain string = `doesNotContain` // 不包含
	OperatorIsEmpty        string = `isEmpty`        // 为空
	OperatorIsNotEmpty     string = `isNotEmpty`     // 不为空
	OperatorIsGreater      string = `isGreater`      // 大于
	OperatorIsGreaterEqual string = `isGreaterEqual` // 大于等于
	OperatorIsLess         string = `isLess`         // 小于
	OperatorIsLessEqual    string = `isLessEqual`    // 小于等于
	OperatorLike           string = `like`           // LIKE
	OperatorIn             string = `in`             // IN
)
