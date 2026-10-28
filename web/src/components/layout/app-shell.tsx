import { useSession } from "@/app/providers/session-provider";
import { ThemeSwitch } from "@/components/theme-switch";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar } from "@/features/profile/components/avatar";
import { perform } from "@/lib/actions";
import { api, json } from "@/lib/api/client";
import {
  ArrowUpRight,
  BookOpen,
  ChevronDown,
  ClipboardList,
  FolderTree,
  LayoutDashboard,
  LogOut,
  Menu,
  Paperclip,
  Settings,
  UserRound,
  Users,
} from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Link, NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
export const adminNav = [
  ["/admin", "总览", LayoutDashboard],
  ["/admin/categories", "作业分类", FolderTree],
  ["/admin/assignments", "作业管理", ClipboardList],
  ["/admin/users", "用户管理", Users],
  ["/admin/attachments", "附件管理", Paperclip],
  ["/admin/settings", "设置", Settings],
] as const;

export function Shell({ admin = false }: { admin?: boolean }) {
  const { user, setUser, title, logo } = useSession();
  const nav = useNavigate();
  const loc = useLocation();
  const [menu, setMenu] = useState(false);
  const [accountOpen, setAccountOpen] = useState(false);
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  useEffect(() => () => clearTimeout(hoverTimer.current), []);
  useEffect(() => setMenu(false), [loc.pathname]);
  if (!user) return null;
  const currentTitle = admin
    ? adminNav.find(([p]) => p != "/admin" && loc.pathname.startsWith(p))?.[1] || "总览"
    : loc.pathname === "/profile"
      ? "个人信息"
      : loc.pathname.startsWith("/assignments/")
        ? "作业详情"
        : "我的作业";
  return (
    <div className="shell">
      {menu && <div className="overlay" onClick={() => setMenu(false)} />}
      <aside className={`sidebar ${menu ? "open" : ""}`}>
        <Link to={admin ? "/admin" : "/"} className="brand">
          <span className="brand-icon">
            {logo ? <img src={logo} alt="" /> : <BookOpen size={22} />}
          </span>
          <span className="brand-name">{title}</span>
        </Link>
        <div className="nav-section">{admin ? "工作空间" : "学习空间"}</div>
        <nav>
          {admin ? (
            adminNav.map(([path, label, Icon]) => (
              <NavLink
                key={path}
                to={path}
                end={path === "/admin"}
                className={({ isActive }) => `nav-link ${isActive ? "active" : ""}`}
              >
                <Icon size={18} />
                {label}
              </NavLink>
            ))
          ) : (
            <>
              <NavLink
                to="/"
                end
                className={({ isActive }) => `nav-link ${isActive ? "active" : ""}`}
              >
                <ClipboardList size={18} />
                我的作业
              </NavLink>
              <NavLink
                to="/profile"
                className={({ isActive }) => `nav-link ${isActive ? "active" : ""}`}
              >
                <UserRound size={18} />
                个人信息
              </NavLink>
            </>
          )}
        </nav>
        <div className="sidebar-foot">
          {user.admin && (
            <Link className="nav-link" to={admin ? "/" : "/admin"}>
              <ArrowUpRight size={18} />
              {admin ? "返回用户前台" : "管理后台"}
            </Link>
          )}
        </div>
      </aside>
      <div className="workspace">
        <header className="topbar">
          <div className="row">
            <Button
              className="mobile-menu"
              variant="ghost"
              size="icon"
              onClick={() => setMenu(!menu)}
              aria-label="打开导航"
            >
              <Menu size={20} />
            </Button>
            <span className="small muted topbar-crumb">
              {admin ? "管理后台" : "学习空间"}
              <span style={{ margin: "0 13px", opacity: 0.4 }}>/</span>
              <span style={{ color: "var(--foreground)" }}>{currentTitle}</span>
            </span>
          </div>
          <div className="topbar-right">
            <span className="small muted hidden sm:inline">
              {new Date().toLocaleDateString("zh-CN", {
                month: "long",
                day: "numeric",
                weekday: "long",
              })}
            </span>
            <ThemeSwitch />
            <div className="subtle-divider" />
            <DropdownMenu open={accountOpen} onOpenChange={setAccountOpen}>
              <DropdownMenuTrigger asChild>
                <button
                  className="account-button"
                  onMouseEnter={() => {
                    hoverTimer.current = setTimeout(() => setAccountOpen(true), 180);
                  }}
                  onMouseLeave={() => clearTimeout(hoverTimer.current)}
                  aria-label="账户菜单"
                >
                  <Avatar user={user} />
                  <span className="account-name">
                    {user.username}
                    <small>{user.admin ? "管理员" : "同学"}</small>
                  </span>
                  <ChevronDown size={14} className="muted" />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                align="end"
                className="account-menu"
                onMouseLeave={() => setAccountOpen(false)}
              >
                <DropdownMenuLabel>{user.username}</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => nav(admin ? `/admin/users/${user.id}` : "/profile")}
                >
                  <UserRound size={15} />
                  查看用户信息
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() =>
                    perform(async () => {
                      await api("/logout", json("POST"));
                      setUser(null);
                      nav("/login");
                    })
                  }
                >
                  <LogOut size={15} />
                  退出登录
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>
        <main className="content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
