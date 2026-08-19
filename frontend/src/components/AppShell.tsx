import {
  BadgeCheck,
  Beaker,
  ClipboardCheck,
  FlaskConical,
  History,
  LogOut,
  Menu,
  ScrollText,
  TestTubes,
  UserRound,
  X,
} from "lucide-react";
import { NavLink, Outlet } from "react-router-dom";
import { useState } from "react";
import { useAuth } from "../state/AuthContext";
import { IconButton } from "./ui";

const links = [
  ["/", "待办工作台", ClipboardCheck],
  ["/samples", "样品与委托", TestTubes],
  ["/custody", "待确认交接", History],
  ["/tasks", "试验任务", FlaskConical],
  ["/reviews", "待复核", BadgeCheck],
  ["/certificates", "证书", ScrollText],
  ["/methods", "方法版本", Beaker],
] as const;
const roleNames = {
  registrar: "登记员",
  technician: "试验员",
  reviewer: "复核员",
  manager: "实验室负责人",
};
export function AppShell() {
  const { user, logout } = useAuth();
  const [mobileOpen, setMobileOpen] = useState(false);
  return (
    <div className="app-shell">
      <header className="topbar">
        <IconButton label="打开导航" onClick={() => setMobileOpen(true)}>
          <Menu size={20} />
        </IconButton>
        <div className="brand-mark">
          <FlaskConical size={20} />
          <strong>材检实验室</strong>
        </div>
        <div className="account">
          <span>
            <b>{user?.display_name}</b>
            <small>{user && roleNames[user.role]}</small>
          </span>
          <NavLink to="/profile" className="icon-button" title="个人会话">
            <UserRound size={18} />
          </NavLink>
          <IconButton label="退出登录" onClick={() => void logout()}>
            <LogOut size={18} />
          </IconButton>
        </div>
      </header>
      <aside className={`sidebar ${mobileOpen ? "sidebar-open" : ""}`}>
        <div className="sidebar-head">
          <span>业务工作区</span>
          <IconButton label="关闭导航" onClick={() => setMobileOpen(false)}>
            <X size={18} />
          </IconButton>
        </div>
        <nav>
          {links.map(([to, label, Icon]) => (
            <NavLink key={to} to={to} onClick={() => setMobileOpen(false)}>
              <Icon size={18} />
              <span>{label}</span>
            </NavLink>
          ))}
        </nav>
        <footer>
          时区 Asia/Shanghai
          <br />
          数据以服务端结果为准
        </footer>
      </aside>
      {mobileOpen && (
        <div className="sidebar-scrim" onClick={() => setMobileOpen(false)} />
      )}
      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
}
