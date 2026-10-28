import { useSession } from "@/app/providers/session-provider";
import { LoadState } from "@/components/feedback/load-state";
import { Field } from "@/components/form/field";
import { PageHeading } from "@/components/page-heading";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useData } from "@/hooks/use-data";
import { perform } from "@/lib/actions";
import { api, json } from "@/lib/api/client";
import type { Settings } from "@/types/models";
import { Save } from "lucide-react";
import { useEffect, useState } from "react";
export function SettingsPage() {
  const q = useData(() => api<Settings>("/admin/settings"));
  const [form, setForm] = useState<Settings>({});
  const [busy, setBusy] = useState(false);
  const { refreshTitle } = useSession();
  useEffect(() => {
    if (q.data) setForm(q.data);
  }, [q.data]);
  const set = (k: string, v: string) => setForm((current) => ({ ...current, [k]: v }));
  const [uploading, setUploading] = useState(false);
  const upload = async (key: string, file: File) => {
    setUploading(true);
    await perform(async () => {
      if (file.size > 2 * 1024 ** 2) throw new Error("图片不能超过 2MB");
      const body = new FormData();
      body.append("image", file);
      const result = await api<{ url: string }>("/admin/settings/images", {
        method: "POST",
        body,
      });
      set(key, result.url);
    });
    setUploading(false);
  };
  return (
    <>
      <PageHeading title="站点设置" description="管理站点信息、上传规则与账号绑定。" />
      <LoadState {...q} retry={q.reload}>
        <form
          className="stack"
          onSubmit={(e) => {
            e.preventDefault();
            setBusy(true);
            perform(async () => {
              await api("/admin/settings", json("PUT", form));
              await refreshTitle();
              q.reload();
            }, "设置已保存").finally(() => setBusy(false));
          }}
        >
          <Panel title="基础设置">
            <div className="form-grid">
              <Field
                label="站点标题"
                hint="显示在登录页和导航栏，支持中文或英文，最多 120 个字符。"
              >
                <Input
                  required
                  disabled={busy}
                  maxLength={120}
                  value={form.site_title || ""}
                  onChange={(e) => set("site_title", e.target.value)}
                />
              </Field>
              <Field label="网站标题" hint="浏览器标签页显示的标题；留空时使用站点标题。">
                <Input
                  maxLength={120}
                  value={form.browser_title || ""}
                  onChange={(e) => set("browser_title", e.target.value)}
                  disabled={busy}
                  placeholder={form.site_title || "作业驿站"}
                />
              </Field>
              {(
                [
                  ["site_logo", "页面 Logo"],
                  ["site_icon", "站点图标（Icon）"],
                ] as const
              ).map(([key, label]) => (
                <div className="stack" key={key}>
                  <Field
                    label={`${label}地址`}
                    hint="填写 HTTP(S) 图片链接或站内路径；留空恢复默认。"
                  >
                    <Input
                      value={form[key] || ""}
                      onChange={(e) => set(key, e.target.value)}
                      disabled={busy || uploading}
                      placeholder="https://example.com/image.png"
                    />
                  </Field>
                  <Field
                    label={`上传${label}`}
                    hint="支持 PNG、JPEG、GIF、WebP，最大 2MB、400 万像素；上传后点击保存设置生效。Icon 建议使用正方形图片。"
                  >
                    <Input
                      type="file"
                      accept="image/png,image/jpeg,image/gif,image/webp"
                      disabled={busy || uploading}
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        e.target.value = "";
                        if (file) void upload(key, file);
                      }}
                    />
                  </Field>
                  {form[key] && (
                    <img
                      className="branding-preview"
                      src={form[key]}
                      alt={`${label}预览`}
                    />
                  )}
                  {form[key] && (
                    <Button
                      type="button"
                      variant="outline"
                      disabled={busy || uploading}
                      onClick={() => set(key, "")}
                    >
                      恢复默认{label}
                    </Button>
                  )}
                </div>
              ))}
              <Field label="用户注册">
                <select
                  value={form.registration_enabled || "true"}
                  onChange={(e) => set("registration_enabled", e.target.value)}
                >
                  <option value="true">允许新用户注册</option>
                  <option value="false">关闭注册</option>
                </select>
              </Field>
            </div>
          </Panel>
          <Panel title="附件与头像">
            <div className="form-grid">
              <Field
                label="默认允许的附件后缀"
                hint="逗号分隔，留空允许全部。此设置用于新作业和用户附件库。"
              >
                <Input
                  value={form.default_extensions || ""}
                  onChange={(e) => set("default_extensions", e.target.value)}
                  placeholder="pdf, zip, docx"
                />
              </Field>
              <Field label="默认作业附件最大大小（MB）">
                <Input
                  type="number"
                  min={0.01}
                  max={10240}
                  step="any"
                  value={+(form.default_max_size || 0) / 1024 ** 2}
                  onChange={(e) =>
                    set(
                      "default_max_size",
                      String(Math.round(+e.target.value * 1024 ** 2)),
                    )
                  }
                />
              </Field>
              <Field
                label="头像上传最大大小（MB）"
                hint="后端会将图片转为 WebP 保存，图片最多 2500 万像素。"
              >
                <Input
                  type="number"
                  min={0.01}
                  max={10240}
                  step="any"
                  value={+(form.avatar_max_size || 0) / 1024 ** 2}
                  onChange={(e) =>
                    set(
                      "avatar_max_size",
                      String(Math.round(+e.target.value * 1024 ** 2)),
                    )
                  }
                />
              </Field>
            </div>
          </Panel>
          <Panel title="实名认证">
            <div className="form-grid">
              <Field
                full
                label="认证 API URL"
                hint="留空关闭实名认证。通过 GET 请求调用，{id} 替换为学号、{name} 替换为姓名，参数自动进行 URL 编码。"
              >
                <Input
                  value={form.realname_api_url || ""}
                  placeholder="https://example.com/verify?id={id}&name={name}"
                  onChange={(e) => set("realname_api_url", e.target.value)}
                />
              </Field>
              <Field
                label="成功判断字段"
                hint="支持嵌套路径与数组下标，例如 success、data.success、results.0.ok。"
              >
                <Input
                  required
                  value={form.realname_success_field || ""}
                  placeholder="success"
                  onChange={(e) => set("realname_success_field", e.target.value)}
                />
              </Field>
              <Field
                label="成功字段的期望值"
                hint={'填写 JSON 值，例如 true、0、200 或 "ok"；字符串必须加双引号。'}
              >
                <Input
                  required
                  value={form.realname_success_value ?? ""}
                  placeholder="true"
                  onChange={(e) => set("realname_success_value", e.target.value)}
                />
              </Field>
              <Field label="返回学号字段" hint="返回的学号仍须与用户输入一致。">
                <Input
                  required
                  value={form.realname_id_field || ""}
                  placeholder="data.id"
                  onChange={(e) => set("realname_id_field", e.target.value)}
                />
              </Field>
              <Field label="返回姓名字段" hint="返回的姓名仍须与用户输入一致。">
                <Input
                  required
                  value={form.realname_name_field || ""}
                  placeholder="data.name"
                  onChange={(e) => set("realname_name_field", e.target.value)}
                />
              </Field>
            </div>
          </Panel>
          <Panel title="QQ 绑定">
            <Field
              label="绑定码 Secret"
              hint="至少 32 个字符，与 NoneBot 插件 HOMEWORK_BINDING_SECRET 保持一致。修改后，尚未使用的旧绑定码将失效。"
            >
              <Input
                type="password"
                autoComplete="off"
                value={form.binding_secret || ""}
                onChange={(e) => set("binding_secret", e.target.value)}
                minLength={32}
                required
              />
            </Field>
          </Panel>
          <div>
            <Button type="submit" disabled={busy || uploading}>
              <Save size={16} />
              {uploading ? "图片上传中…" : busy ? "保存中…" : "保存设置"}
            </Button>
          </div>
        </form>
      </LoadState>
    </>
  );
}
