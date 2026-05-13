---
name: cloud-ticket-approval
description: 云平台工单审批技能，支持查看、同意、拒绝审批工单，支持自动审批
---

# 云平台工单审批

## 工具说明

### list_pending_tickets
查询当前用户相关的工单列表。

**参数：**
- `status` (可选): 工单状态过滤，可选值: pending/approved/rejected，默认 pending
- `filter` (可选): 过滤类型，可选值: my_apply(我发起的), my_approve(我审批的), 空(全部相关)
- `page` (可选): 页码，从1开始
- `page_size` (可选): 每页数量，最大100

### get_ticket_detail
获取工单详细信息，审批前必须先查看详情。

**参数：**
- `ticket_id` (必填): 工单ID，格式如 T-20260514-001

### review_ticket
审批工单，同意或拒绝。

**参数：**
- `ticket_id` (必填): 工单ID
- `action` (必填): 审批动作，可选值: agree(同意) / disagree(拒绝)
- `comment` (拒绝时必填): 审批意见

## 审批流程规范

1. **必须先查后审**
   - 调用 review_ticket 前，必须先调用 get_ticket_detail
   - 确认工单内容合理后再操作

2. **安全确认机制**
   - 只能审批分配给自己的工单
   - 如果工单的 approver 不是当前用户，必须拒绝操作并提示无权操作

3. **拒绝必须说明原因**
   - action=disagree 时，comment 为必填
   - 原因必须具体明确，便于申请人理解

## 使用示例

### 场景：查看待审批工单

```
list_pending_tickets(status="pending", filter="my_approve")
```

### 场景：查看我发起的工单

```
list_pending_tickets(filter="my_apply")
```

### 场景：查看工单详情并审批

1. 查看详情
   ```
   get_ticket_detail(ticket_id="T-20260514-001")
   ```

2. 同意工单
   ```
   review_ticket(ticket_id="T-20260514-001", action="agree", comment="同意")
   ```

### 场景：拒绝工单

1. 查看详情后发现不合理
   ```
   get_ticket_detail(ticket_id="T-20260514-002")
   ```

2. 拒绝并说明原因
   ```
   review_ticket(ticket_id="T-20260514-002", action="disagree", comment="申请信息不完整，请补充后重新提交")
   ```

## 自动审批

支持用户自定义规则进行自动审批。

### 使用方式

1. 复制模板文件 `skill/auto-approval-template.md` 为 `skill/my-auto-approval.md`
2. 修改模板中的规则，定义自己的审批条件
3. 对Claude说："按照我的自动审批规则执行"

### 规则格式

```markdown
### 规则名称

**匹配条件：**
- 场景 = "xxx"
- 标题包含 "xxx"
- 申请人 = "xxx"

**执行动作：**
- 动作：agree 或 disagree
- 原因：xxx
```

### 支持的匹配条件

| 条件 | 说明 | 示例 |
|-----|------|------|
| 场景 | 工单的scene字段 | 场景 = "创建集群" |
| 标题 | 工单标题 | 标题包含 "测试" |
| 申请人 | 工单的applicant字段 | 申请人 = "zhangsan" |
| 审批人 | 工单的approver字段 | 审批人 = "当前用户" |

详细模板见 `skill/auto-approval-template.md`
