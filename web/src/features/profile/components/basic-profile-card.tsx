import { useSession } from "@/app/providers/session-provider";
import { Field } from "@/components/form/field";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Avatar } from "@/features/profile/components/avatar";
import { api, json } from "@/lib/api/client";
import { dateText } from "@/lib/format";
import type { User } from "@/types/models";
import { Save, Upload } from "lucide-react";
import { useState } from "react";
import { useProfileAction } from "../hooks/use-profile-action";
export function BasicProfileCard() {
  const { user, setUser } = useSession();
  const { busy, act } = useProfileAction();
  const [username, setUsername] = useState(user?.username || "");
  if (!user) return null;
  return (
    <Panel title="基本资料">
      <div className="profile-summary">
        <Avatar user={user} />
        <div>
          <h2>{user.username}</h2>
          <p className="small muted" style={{ marginTop: 7 }}>
            注册于 {dateText(user.created_at)} · {user.admin ? "管理员" : "同学"}
          </p>
        </div>
        <label style={{ marginLeft: "auto" }}>
          <Button type="button" variant="outline" size="sm" disabled={!!busy} asChild>
            <span>
              <Upload size={14} />
              上传头像
            </span>
          </Button>
          <input
            aria-label="上传头像"
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp"
            hidden
            disabled={!!busy}
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) {
                const data = new FormData();
                data.append("avatar", file);
                act(
                  "avatar",
                  async () => {
                    setUser(
                      await api<User>("/me/avatar", {
                        method: "POST",
                        body: data,
                      }),
                    );
                  },
                  "头像已更新",
                );
              }
              e.target.value = "";
            }}
          />
        </label>
      </div>
      <form
        className="form-grid"
        onSubmit={(e) => {
          e.preventDefault();
          act(
            "name",
            async () => setUser(await api<User>("/me", json("PATCH", { username }))),
            "资料已保存",
          );
        }}
      >
        <Field label="用户名" hint="2–32 位文字、数字或 _.-">
          <Input
            required
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            maxLength={32}
          />
        </Field>
        <div
          style={{
            display: "flex",
            alignItems: "flex-start",
            paddingTop: 28,
          }}
        >
          <Button type="submit" disabled={!!busy}>
            <Save size={14} />
            {busy === "name" ? "保存中…" : "保存资料"}
          </Button>
        </div>
      </form>
    </Panel>
  );
}
