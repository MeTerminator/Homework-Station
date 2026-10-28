import { useSession } from "@/app/providers/session-provider";
import { BindingInstructions } from "@/components/binding-instructions";
import { Field } from "@/components/form/field";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { api, json } from "@/lib/api/client";
import type { User } from "@/types/models";
import { CheckCircle2, UserRound } from "lucide-react";
import { useState } from "react";
import { useProfileAction } from "../hooks/use-profile-action";
export function QQBindingCard() {
  const { user, setUser, qqBindingInstructions } = useSession();
  const { busy, act } = useProfileAction();
  const [code, setCode] = useState("");
  if (!user) return null;
  return (
    <Panel
      title="QQ 绑定"
      action={<UserRound size={20} style={{ color: "var(--primary)" }} />}
    >
      {user.qq ? (
        <>
          <div className="note row">
            <CheckCircle2 size={18} />
            已绑定 QQ
          </div>
          <p style={{ marginTop: 25, fontSize: 22, fontWeight: 600 }}>{user.qq}</p>
          <p className="small muted" style={{ marginTop: 16 }}>
            如需更换，请联系管理员解绑。
          </p>
        </>
      ) : (
        <form
          className="stack"
          onSubmit={(e) => {
            e.preventDefault();
            act(
              "qq",
              async () => {
                setUser(await api<User>("/me/bind/qq", json("POST", { code })));
                setCode("");
              },
              "QQ 绑定成功",
            );
          }}
        >
          <BindingInstructions text={qqBindingInstructions} />
          <Field label="绑定码">
            <Textarea
              required
              rows={5}
              value={code}
              onChange={(e) => setCode(e.target.value.trim())}
              placeholder="粘贴机器人发送的绑定码"
              style={{ fontFamily: "monospace", fontSize: 12 }}
            />
          </Field>
          <div>
            <Button type="submit" disabled={!!busy}>
              {busy === "qq" ? "验证中…" : "绑定 QQ"}
            </Button>
          </div>
        </form>
      )}
    </Panel>
  );
}
