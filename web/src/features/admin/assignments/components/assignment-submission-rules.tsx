import { CheckField } from "@/components/form/check-field";
import { Field } from "@/components/form/field";
import { Panel } from "@/components/panel";
import { Input } from "@/components/ui/input";
import type { Assignment } from "@/types/models";
import type { AssignmentFieldChange } from "../types";
export function AssignmentSubmissionRules({
  form,
  onChangeField,
}: {
  form: Assignment;
  onChangeField: AssignmentFieldChange;
}) {
  return (
    <Panel title="提交规则">
      <div className="form-grid">
        <Field label="提交类型">
          <select
            value={form.allow_attachments ? "attachment" : "text"}
            onChange={(e) =>
              onChangeField("allow_attachments", e.target.value === "attachment")
            }
          >
            <option value="attachment">允许文本和附件</option>
            <option value="text">仅纯文本</option>
          </select>
        </Field>
        <Field label="单个附件最大大小（MB）">
          <Input
            type="number"
            min={0.01}
            max={10240}
            step="any"
            value={form.max_size / 1024 ** 2}
            onChange={(e) =>
              onChangeField("max_size", Math.round(+e.target.value * 1024 ** 2))
            }
          />
        </Field>
        <Field label="允许的附件后缀" hint="用逗号分隔，如 pdf,zip,docx。留空允许全部。">
          <Input
            value={form.extensions}
            onChange={(e) => onChangeField("extensions", e.target.value)}
            placeholder="允许全部文件类型"
          />
        </Field>
        <div className="stack" style={{ gap: 18, paddingTop: 24 }}>
          <CheckField
            label="必须完成实名认证后才能提交"
            checked={form.require_real_name}
            onChange={(v) => onChangeField("require_real_name", v)}
          />
          <CheckField
            label="必须绑定 QQ 后才能提交"
            checked={form.require_qq}
            onChange={(v) => onChangeField("require_qq", v)}
          />
          <CheckField
            label="允许多次提交（保留全部历史）"
            checked={form.multiple}
            onChange={(v) => onChangeField("multiple", v)}
          />
          <CheckField
            label="允许用户撤回作业"
            checked={form.withdraw}
            onChange={(v) => onChangeField("withdraw", v)}
          />
        </div>
      </div>
    </Panel>
  );
}
