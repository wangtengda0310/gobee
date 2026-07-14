/**
 * 通用树形右键菜单功能
 *
 * 提取战斗测试和配表测试中复用的右键菜单功能：
 * - 责任人归一：将首个有负责人的子节点同步到同组所有子节点
 * - 执行分类：执行指定分组下的所有子节点
 */
import type {DropdownOption} from "naive-ui";
import type {DialogApiInjection} from "naive-ui/es/dialog/src/DialogProvider";
import type {MessageApiInjection} from "naive-ui/es/message/src/MessageProvider";

/** 负责人字段类型 */
export type ManagerField = "QA" | "Designer" | "Programmer";

/** 负责人列表结构 */
export type ManagerList = {
    QA: string[];
    Designer: string[];
    Programmer: string[];
};

/** 子节点类型，需要包含 label 和负责人字段
 * 字段值可以是 string[]（如 QA/Designer/Programmer）或 string（如 caseManager）
 */
export type ManagerChild = {
    label?: string;
    modified?: boolean;
    [key: string]: any;
};

/**
 * 创建责任人归一下拉菜单项
 * 支持 string[] 类型（如 QA/Designer/Programmer）和 string 类型（如 caseManager）
 * @param dialog 对话框 API
 * @param message 消息 API
 * @param children 子节点数组
 * @param field 要归一的负责人字段
 * @param onSync 同步后的回调（可选）
 */
export function createManagerToOneOption<T extends ManagerChild>(
    dialog: DialogApiInjection,
    message: MessageApiInjection,
    children: T[] | undefined,
    field: ManagerField | string,
    onSync?: () => void,
): DropdownOption {
    const fieldLabelMap: Record<string, string> = {
        QA: "QA",
        Designer: "策划",
        Programmer: "程序",
        caseManager: "负责人",
    };
    const label = fieldLabelMap[field] ?? field;
    return {
        title: `${label}归一`,
        key: `manager-to-one-${String(field).toLowerCase()}`,
        props: {
            onClick() {
                dialog.warning({
                    title: `${label}归一`,
                    content: `该操作会将该组里顺序检索到的首个${label}同步到组下的所有子项`,
                    positiveText: "确定",
                    negativeText: "取消",
                    draggable: true,
                    onPositiveClick: () => {
                        const hasManagerChild = children?.find((c) => {
                            const val = c[field];
                            if (Array.isArray(val)) {
                                return val.length > 0;
                            }
                            return typeof val === "string" && val !== "";
                        });
                        if (hasManagerChild) {
                            const sourceVal = hasManagerChild[field];
                            const isArrayField = Array.isArray(sourceVal);
                            children?.forEach((c) => {
                                const targetVal = c[field];
                                if (isArrayField) {
                                    // 数组类型字段（如 QA/Designer/Programmer）
                                    const sourceList = sourceVal as string[];
                                    if (
                                        !targetVal ||
                                        !Array.isArray(targetVal) ||
                                        JSON.stringify(targetVal) !== JSON.stringify(sourceList)
                                    ) {
                                        (c as any)[field] = [...sourceList];
                                        c.modified = true;
                                    }
                                } else {
                                    // 字符串类型字段（如 caseManager）
                                    const sourceStr = sourceVal as string;
                                    if (targetVal !== sourceStr) {
                                        (c as any)[field] = sourceStr;
                                        c.modified = true;
                                    }
                                }
                            });
                            onSync?.();
                        } else {
                            message.error(`该组不存在含有${label}的子项`);
                        }
                    },
                    onNegativeClick: () => {},
                });
            },
        },
    };
}

/**
 * 创建执行分类下拉菜单项
 * @param message 消息 API
 * @param label 分组名称（用于提示）
 * @param onExec 执行回调函数
 * @param key 菜单项 key
 * @param title 菜单显示标题，默认 "执行分类"
 */
export function createExecCateOption(
    message: MessageApiInjection,
    label: string | undefined,
    onExec: () => void,
    key: string = "exec-cate",
    title: string = "执行分类",
): DropdownOption {
    return {
        title: title,
        key: key,
        props: {
            onClick() {
                if (!label) {
                    message.error("分组名称未找到");
                    return;
                }
                onExec();
            },
        },
    };
}

/**
 * 创建分隔线菜单项
 */
export function createDividerOption(): DropdownOption {
    return {
        type: "divider",
        key: "divider",
    };
}
