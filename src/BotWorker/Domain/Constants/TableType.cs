using System.ComponentModel;

namespace BotWorker.Core.Data
{
    /// <summary>
    /// simpleTable ��չʾ����������
    /// </summary>
    public enum TableType
    {
        /// <summary>
        /// ��������
        /// </summary>
        [Description("BasicTable")]
        Basic,

        /// <summary>
        /// ��������
        /// </summary>
        [Description("ReportTable")]
        Report,

        /// <summary>
        /// ��������
        /// </summary>
        [Description("OtherTable")]
        Other

    }
}


