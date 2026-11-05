import type { ThemeConfig } from "antd";

// 高纯度蓝只用于动作与选中态，中性色负责承载高密度业务信息。
export const neutralPreviewTheme: ThemeConfig = {
  token: {
    borderRadius: 6,
    borderRadiusLG: 8,
    borderRadiusSM: 4,
    controlHeight: 32,
    fontSize: 14,
    fontFamily:
      '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "Helvetica Neue", Helvetica, Arial, sans-serif',
    fontWeightStrong: 600,
    motionDurationMid: "0.18s",
    colorPrimary: "#2F63E9",
    colorInfo: "#2F63E9",
    colorSuccess: "#16805C",
    colorWarning: "#B56B18",
    colorError: "#D1434B",
    colorText: "#20232A",
    colorTextSecondary: "#697180",
    colorBorder: "#DDE2E9",
    colorBgLayout: "#F4F6F8",
    // 侧栏会占用桌面宽度，1400px 起才恢复全部表格列。
    screenXXL: 1400,
    screenXXLMin: 1400,
    screenXLMax: 1399,
  },
  components: {
    Button: { controlHeight: 32, borderRadius: 4 },
    Input: { controlHeight: 32, borderRadius: 4 },
    InputNumber: { controlHeight: 32, borderRadius: 4 },
    Select: { controlHeight: 32, borderRadius: 4 },
    Menu: {
      itemBorderRadius: 5,
      itemSelectedBg: "#EAF0FF",
      itemSelectedColor: "#2457D0",
      itemHoverBg: "#F1F4F9",
      subMenuItemBg: "transparent",
    },
    Table: {
      headerBg: "#FFFFFF",
      headerColor: "#1F2329",
      borderColor: "#D9DDE3",
      rowHoverBg: "#F5F8FF",
      headerBorderRadius: 0,
      cellPaddingBlock: 0,
      cellPaddingInline: 16,
    },
    Tag: { defaultBg: "#F1F3F6", defaultColor: "#596273" },
    Modal: { borderRadiusLG: 8 },
    Drawer: { paddingLG: 20 },
  },
};
