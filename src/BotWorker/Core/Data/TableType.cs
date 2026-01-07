using System.ComponentModel;

namespace sz84.Core.Data
{
    /// <summary>
    /// simpleTable 可展示的数据类型
    /// </summary>
    public enum TableType
    {
        /// <summary>
        /// 基础数据
        /// </summary>
        [Description("BasicTable")]
        Basic,

        /// <summary>
        /// 报表数据
        /// </summary>
        [Description("ReportTable")]
        Report,

        /// <summary>
        /// 其它数据
        /// </summary>
        [Description("OtherTable")]
        Other

    }
}
