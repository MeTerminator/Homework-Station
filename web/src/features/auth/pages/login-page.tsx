import { useSession } from "@/app/providers/session-provider";
import { Field } from "@/components/form/field";
import { ThemeSwitch } from "@/components/theme-switch";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useData } from "@/hooks/use-data";
import { api, json } from "@/lib/api/client";
import type { User } from "@/types/models";
import { ArrowRight, BookOpen, GraduationCap } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
export function AuthPage() {
  const { user, setUser, title } = useSession();
  const nav = useNavigate();
  const [register, setRegister] = useState(false);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [captcha, setCaptcha] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const cap = useData(() => api<{ id: string; image: string }>("/captcha"), [register]);
  const site = useData(() => api<{ registration_enabled: string }>("/public-settings"));
  useEffect(() => {
    if (user) nav("/", { replace: true });
  }, [user, nav]);
  return (
    <div className="auth-page">
      <div className="auth-art">
        <div className="brand">
          <span className="brand-icon">
            <BookOpen size={22} />
          </span>
          {title}
        </div>
        <div className="orb" />
        <div className="auth-quote">
          <h1>作业提交与收集</h1>
          <p>查看作业要求、提交作业并保留历史记录。</p>
        </div>
      </div>
      <div className="auth-main">
        <div className="auth-theme">
          <ThemeSwitch />
        </div>
        <div className="auth-box">
          <span className="stat-icon" style={{ marginBottom: 24 }}>
            <GraduationCap size={22} />
          </span>
          <h1>{register ? "创建你的账号" : "欢迎回来"}</h1>
          <p className="muted" style={{ marginTop: 12 }}>
            {register ? "填写账号信息完成注册。" : "登录后查看和提交作业。"}
          </p>
          <form
            className="auth-form"
            onSubmit={(e) => {
              e.preventDefault();
              setError("");
              setBusy(true);
              api<User>(
                register ? "/register" : "/login",
                json("POST", {
                  username,
                  password,
                  confirm,
                  captcha,
                  captcha_id: cap.data?.id,
                }),
              )
                .then((u) => {
                  setUser(u);
                  nav("/", { replace: true });
                })
                .catch((e) => {
                  setError(e.message);
                  if (register) {
                    cap.reload();
                    setCaptcha("");
                  }
                })
                .finally(() => setBusy(false));
            }}
          >
            {error && (
              <div className="error-box" role="alert">
                {error}
              </div>
            )}
            <Field label="用户名">
              <Input
                aria-label="用户名"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                autoComplete="username"
                placeholder="请输入用户名"
                maxLength={32}
              />
            </Field>
            <Field label="密码">
              <Input
                aria-label="密码"
                required
                type="password"
                minLength={6}
                maxLength={72}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete={register ? "new-password" : "current-password"}
                placeholder="请输入密码"
              />
            </Field>
            {register && (
              <>
                <Field label="确认密码">
                  <Input
                    aria-label="确认密码"
                    required
                    type="password"
                    minLength={6}
                    maxLength={72}
                    value={confirm}
                    onChange={(e) => setConfirm(e.target.value)}
                    autoComplete="new-password"
                    placeholder="请再次输入密码"
                  />
                </Field>
                <Field label="图形验证码" hint="点击图片可刷新，验证码 5 分钟内有效。">
                  <div className="auth-captcha">
                    <Input
                      aria-label="图形验证码"
                      required
                      pattern="[0-9]{4}"
                      inputMode="numeric"
                      maxLength={4}
                      value={captcha}
                      onChange={(e) => setCaptcha(e.target.value)}
                      placeholder="4 位数字"
                    />
                    <button
                      type="button"
                      onClick={cap.reload}
                      aria-label="刷新验证码"
                      style={{ background: "none", border: 0, padding: 0 }}
                    >
                      {cap.data ? (
                        <img src={cap.data.image} alt="4 位数字验证码" />
                      ) : (
                        <span className="small link">
                          {cap.error ? "加载失败，点击重试" : "加载中…"}
                        </span>
                      )}
                    </button>
                  </div>
                </Field>
              </>
            )}
            <Button
              type="submit"
              size="lg"
              disabled={busy || (register && !cap.data)}
              style={{ marginTop: 4 }}
            >
              {busy ? "处理中…" : register ? "创建账号" : "登 录"}
              <ArrowRight size={16} />
            </Button>
          </form>
          <p className="auth-switch">
            {register ? "已有账号？" : "还没有账号？"}{" "}
            <button
              className="link"
              style={{ border: 0, background: "none" }}
              disabled={!register && site.data?.registration_enabled === "false"}
              onClick={() => {
                setRegister(!register);
                setError("");
                setCaptcha("");
              }}
            >
              {register
                ? "立即登录"
                : site.data?.registration_enabled === "false"
                  ? "注册已关闭"
                  : "立即注册"}
            </button>
          </p>
        </div>
      </div>
    </div>
  );
}
