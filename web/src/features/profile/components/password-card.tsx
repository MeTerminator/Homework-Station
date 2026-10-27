import { useSession } from "@/app/providers/session-provider";
import { Field } from "@/components/form/field";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api, json } from "@/lib/api/client";
import { KeyRound } from "lucide-react";
import { useState } from "react";
import { useProfileAction } from "../hooks/use-profile-action";
export function PasswordCard() {
  const { user } = useSession();
  const { busy, act } = useProfileAction();
  const [pw, setPW] = useState({ old: "", password: "", confirm: "" });
  if (!user) return null;
  return (
    <Panel
      title="修改密码"
      action={<KeyRound size={20} style={{ color: "var(--primary)" }} />}
    >
      <form
        className="form-grid"
        onSubmit={(e) => {
          e.preventDefault();
          act(
            "password",
            async () => {
              await api("/me/password", json("POST", pw));
              setPW({ old: "", password: "", confirm: "" });
            },
            "密码已修改",
          );
        }}
      >
        <Field label="旧密码">
          <Input
            required
            type="password"
            autoComplete="current-password"
            value={pw.old}
            onChange={(e) => setPW({ ...pw, old: e.target.value })}
          />
        </Field>
        <div />
        <Field label="新密码">
          <Input
            required
            type="password"
            autoComplete="new-password"
            minLength={6}
            maxLength={72}
            value={pw.password}
            onChange={(e) => setPW({ ...pw, password: e.target.value })}
          />
        </Field>
        <Field label="确认新密码">
          <Input
            required
            type="password"
            autoComplete="new-password"
            minLength={6}
            maxLength={72}
            value={pw.confirm}
            onChange={(e) => setPW({ ...pw, confirm: e.target.value })}
          />
        </Field>
        <div>
          <Button type="submit" disabled={!!busy}>
            {busy === "password" ? "保存中…" : "修改密码"}
          </Button>
        </div>
      </form>
    </Panel>
  );
}
