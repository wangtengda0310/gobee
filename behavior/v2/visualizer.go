// Package behavior 提供行为树的可视化导出功能。
// 支持 DOT 格式导出，可使用 Graphviz 渲染。
package behavior

import (
	"fmt"
	"io"
	"strings"
)

// VisualizableNode 是可提供可视化名称的节点接口。
// 自定义节点可实现此接口以在可视化时显示自定义名称。
type VisualizableNode interface {
	Node
	// VisualName 返回节点在可视化图中的显示名称。
	VisualName() string
}

// ExportDOT 将行为树导出为 DOT 格式。
// DOT 格式可被 Graphviz 工具渲染为图形。
//
// 参数:
//   - root: 行为树的根节点
//   - writer: 用于写入 DOT 内容的 io.Writer
//
// 返回值:
//   - error: 写入过程中的错误
//
// 渲染方法:
//
//	dot -Tpng tree.dot -o tree.png
//	dot -Tsvg tree.dot -o tree.svg
//
// 示例:
//
//	var buf bytes.Buffer
//	err := ExportDOT(tree, &buf)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(buf.String())
func ExportDOT(root Node, writer io.Writer) error {
	return exportDOT(root, writer, "behavior_tree")
}

func exportDOT(node Node, writer io.Writer, name string) error {
	_, err := fmt.Fprintf(writer, "digraph %s {\n", name)
	if err != nil {
		return err
	}

	// 设置全局节点样式
	_, err = fmt.Fprintln(writer, `  node [shape=box, style="rounded, filled", fontname="Arial"];`)
	if err != nil {
		return err
	}

	// 设置全局边样式
	_, err = fmt.Fprintln(writer, `  edge [fontname="Arial"];`)
	if err != nil {
		return err
	}

	nodeID := 0
	err = exportNodeDOT(node, writer, &nodeID, "")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, "}")
	return err
}

func exportNodeDOT(node Node, writer io.Writer, nodeID *int, parentID string) error {
	currentID := fmt.Sprintf("n%d", *nodeID)
	*nodeID++

	// 获取节点信息
	nodeName, nodeColor, nodeShape := getNodeInfo(node)

	// 写入节点定义
	_, err := fmt.Fprintf(writer, `  %s [label="%s", fillcolor="%s", shape="%s"];`+"\n",
		currentID, nodeName, nodeColor, nodeShape)
	if err != nil {
		return err
	}

	// 写入从父节点到当前节点的边
	if parentID != "" {
		_, err = fmt.Fprintf(writer, "  %s -> %s;\n", parentID, currentID)
		if err != nil {
			return err
		}
	}

	// 处理复合节点的子节点
	switch n := node.(type) {
	case *Sequence:
		for _, child := range n.Children {
			err = exportNodeDOT(child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	case *Selector:
		for _, child := range n.Children {
			err = exportNodeDOT(child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	case *Parallel:
		for _, child := range n.Children {
			err = exportNodeDOT(child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	case *RandomSelector:
		for _, child := range n.children {
			err = exportNodeDOT(child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	}

	// 处理装饰器节点的子节点
	switch n := node.(type) {
	case *Inverter:
		if n.child != nil {
			err = exportNodeDOT(n.child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	case *Repeater:
		if n.child != nil {
			err = exportNodeDOT(n.child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	case *Retry:
		if n.child != nil {
			err = exportNodeDOT(n.child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	case *Timeout:
		if n.child != nil {
			err = exportNodeDOT(n.child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	case *Delay:
		if n.child != nil {
			err = exportNodeDOT(n.child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	case *Limiter:
		if n.child != nil {
			err = exportNodeDOT(n.child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	case *UntilSuccess:
		if n.child != nil {
			err = exportNodeDOT(n.child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	case *UntilFailure:
		if n.child != nil {
			err = exportNodeDOT(n.child, writer, nodeID, currentID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// getNodeInfo 获取节点的可视化信息。
// 返回节点名称、填充颜色和形状。
func getNodeInfo(node Node) (name, color, shape string) {
	switch n := node.(type) {
	case *Sequence:
		return "Sequence", "#90EE90", "box" // 浅绿色
	case *Selector:
		return "Selector", "#87CEEB", "box" // 浅蓝色
	case *Parallel:
		return fmt.Sprintf("Parallel\\n(s:%d, f:%d)", n.SuccessPolicy, n.FailurePolicy), "#DDA0DD", "box" // 紫色
	case *RandomSelector:
		return "RandomSelector", "#98FB98", "box" // 淡绿色
	case *Condition:
		return "Condition", "#FFD700", "diamond" // 金色
	case *Action:
		return "Action", "#FFA07A", "ellipse" // 浅鲑鱼色
	case *Inverter:
		return "Inverter", "#F0E68C", "hexagon" // 卡其色
	case *Repeater:
		if n.times < 0 {
			return "Repeater\\n(∞)", "#F0E68C", "hexagon"
		}
		return fmt.Sprintf("Repeater\\n(%d)", n.times), "#F0E68C", "hexagon"
	case *Retry:
		return fmt.Sprintf("Retry\\n(%d)", n.maxTries), "#F0E68C", "hexagon"
	case *Timeout:
		return fmt.Sprintf("Timeout\\n(%v)", n.duration), "#F0E68C", "hexagon"
	case *Delay:
		return fmt.Sprintf("Delay\\n(%d)", n.delayTicks), "#F0E68C", "hexagon"
	case *Limiter:
		return fmt.Sprintf("Limiter\\n(%d)", n.maxCalls), "#F0E68C", "hexagon"
	case *UntilSuccess:
		return "UntilSuccess", "#F0E68C", "hexagon"
	case *UntilFailure:
		return "UntilFailure", "#F0E68C", "hexagon"
	default:
		// 检查是否实现了 VisualizableNode 接口
		if vn, ok := node.(VisualizableNode); ok {
			return vn.VisualName(), "#FFFFFF", "box"
		}
		return "Unknown", "#FFFFFF", "box"
	}
}

// ExportDOTString 将行为树导出为 DOT 格式字符串。
//
// 参数:
//   - root: 行为树的根节点
//
// 返回值:
//   - string: DOT 格式的字符串
//   - error: 导出过程中的错误
//
// 示例:
//
//	dot, err := ExportDOTString(tree)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(dot)
//	// 保存到文件并渲染: echo "$dot" > tree.dot && dot -Tpng tree.dot -o tree.png
func ExportDOTString(root Node) (string, error) {
	var sb strings.Builder
	err := ExportDOT(root, &sb)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}
