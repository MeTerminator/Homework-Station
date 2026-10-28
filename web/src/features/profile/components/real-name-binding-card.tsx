import { useSession } from "@/app/providers/session-provider";
import { Field } from "@/components/form/field";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useData } from "@/hooks/use-data";
import { api, json } from "@/lib/api/client";
import type { User } from "@/types/models";
import { CheckCircle2, ShieldCheck } from "lucide-react";
import { useState } from "react";
import { useProfileAction } from "../hooks/use-profile-action";
export function RealNameBindingCard() {
  const { user, setUser } = useSession();
  const { busy, act } = useProfileAction();
  const identitySettings = useData(() =>
    api<{ realname_enabled: boolean; realname_student_no_placeholder: string }>(
      "/public-settings",
    ),
  );
  const [real, setReal] = useState({ student_no: "", name: "" });
  if (!user) return null;
  return (
    <Panel
      title="实名认证"
      action={<ShieldCheck size={20} style={{ color: "var(--primary)" }} />}
    >
      {user.student_no ? (
        <>
          <div className="note row">
            <CheckCircle2 size={18} />
            已完成实名认证
          </div>
          <div className="stack" style={{ gap: 15, marginTop: 22 }}>
            <span className="small muted">
              姓名{" "}
              <strong style={{ color: "var(--foreground)", marginLeft: 20 }}>
                {user.real_name}
              </strong>
            </span>
            <span className="small muted">
              学号{" "}
              <strong style={{ color: "var(--foreground)", marginLeft: 20 }}>
                {user.student_no}
              </strong>
            </span>
            <span className="small muted">如需修改，请联系管理员解绑。</span>
          </div>
        </>
      ) : identitySettings.loading ? (
        <p className="muted small">正在加载实名认证设置…</p>
      ) : identitySettings.error ? (
        <div className="row">
          <span className="muted small">无法读取实名认证设置</span>
          <Button variant="outline" size="sm" onClick={identitySettings.reload}>
            重试
          </Button>
        </div>
      ) : !identitySettings.data?.realname_enabled ? (
        <div className="note">站点暂未启用实名认证，请联系管理员。</div>
      ) : (
        <form
          className="stack"
          onSubmit={(e) => {
            e.preventDefault();
            act(
              "real",
              async () => setUser(await api<User>("/me/bind/real", json("POST", real))),
              "实名认证成功",
            );
          }}
        >
          <p className="muted small">
            填写学号与真实姓名，通过站点配置的认证服务校验后绑定。
          </p>
          <Field label="学号">
            <Input
              required
              maxLength={128}
              title={identitySettings.data?.realname_student_no_placeholder}
              value={real.student_no}
              onChange={(e) => setReal({ ...real, student_no: e.target.value })}
              placeholder={identitySettings.data?.realname_student_no_placeholder}
            />
          </Field>
          <Field label="姓名">
            <Input
              required
              value={real.name}
              onChange={(e) => setReal({ ...real, name: e.target.value })}
              placeholder="请输入真实姓名"
            />
          </Field>
          <div>
            <Button type="submit" disabled={!!busy}>
              {busy === "real" ? "验证中…" : "验证并绑定"}
            </Button>
          </div>
        </form>
      )}
    </Panel>
  );
}
