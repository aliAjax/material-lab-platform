import {
  AlertTriangle,
  ChevronLeft,
  ChevronRight,
  LoaderCircle,
  RefreshCw,
  Search,
  X,
} from "lucide-react";
import { forwardRef, type ReactNode, useEffect, useRef } from "react";
import type { StatusTone } from "../types/domain";

export function StatusBadge({
  value,
  tone = "neutral",
}: {
  value: string;
  tone?: StatusTone;
}) {
  return <span className={`badge badge-${tone}`}>{value}</span>;
}
export const Button = forwardRef<
  HTMLButtonElement,
  React.ButtonHTMLAttributes<HTMLButtonElement> & {
    variant?: "primary" | "secondary" | "danger" | "ghost";
    icon?: ReactNode;
  }
>(function Button({ children, variant = "primary", icon, ...props }, ref) {
  return (
    <button ref={ref} className={`button button-${variant}`} {...props}>
      {icon}
      {children}
    </button>
  );
});
export function IconButton({
  label,
  children,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & {
  label: string;
  children: ReactNode;
}) {
  return (
    <button className="icon-button" title={label} aria-label={label} {...props}>
      {children}
    </button>
  );
}
export function PageHeader({
  title,
  description,
  actions,
}: {
  title: string;
  description?: string;
  actions?: ReactNode;
}) {
  return (
    <header className="page-header">
      <div>
        <h1>{title}</h1>
        {description && <p>{description}</p>}
      </div>
      <div className="page-actions">{actions}</div>
    </header>
  );
}
export function Field({
  label,
  error,
  hint,
  required,
  children,
}: {
  label: string;
  error?: string;
  hint?: string;
  required?: boolean;
  children: ReactNode;
}) {
  return (
    <label className="field">
      <span>
        {label}
        {required && <b aria-hidden="true"> *</b>}
      </span>
      {children}
      {error ? (
        <small className="field-error">{error}</small>
      ) : (
        hint && <small>{hint}</small>
      )}
    </label>
  );
}
export function EmptyState({
  title = "暂无记录",
  description = "调整筛选条件或创建一条新记录。",
}: {
  title?: string;
  description?: string;
}) {
  return (
    <div className="empty-state">
      <Search size={24} />
      <strong>{title}</strong>
      <span>{description}</span>
    </div>
  );
}
export function LoadingState({ label = "正在加载" }: { label?: string }) {
  return (
    <div className="loading-state">
      <LoaderCircle className="spin" size={22} />
      {label}
    </div>
  );
}
export function ErrorState({
  error,
  retry,
}: {
  error: unknown;
  retry?: () => void;
}) {
  return (
    <div className="error-state" role="alert">
      <AlertTriangle size={22} />
      <div>
        <strong>无法加载数据</strong>
        <p>{error instanceof Error ? error.message : "发生未知错误"}</p>
      </div>
      {retry && (
        <IconButton label="重试" onClick={retry}>
          <RefreshCw size={18} />
        </IconButton>
      )}
    </div>
  );
}
export function Pagination({
  previous,
  next,
  onChange,
}: {
  previous?: string;
  next?: string;
  onChange: (cursor?: string) => void;
}) {
  return (
    <nav className="pagination" aria-label="分页">
      <Button
        variant="secondary"
        disabled={!previous}
        onClick={() => onChange(previous)}
        icon={<ChevronLeft size={16} />}
      >
        上一页
      </Button>
      <Button
        variant="secondary"
        disabled={!next}
        onClick={() => onChange(next)}
      >
        下一页
        <ChevronRight size={16} />
      </Button>
    </nav>
  );
}
export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel = "确认",
  danger,
  onConfirm,
  onClose,
}: {
  open: boolean;
  title: string;
  description: string;
  confirmLabel?: string;
  danger?: boolean;
  onConfirm(): void;
  onClose(): void;
}) {
  const confirmRef = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    if (open) confirmRef.current?.focus();
  }, [open]);
  if (!open) return null;
  return (
    <div
      className="dialog-backdrop"
      role="presentation"
      onMouseDown={(event) => event.target === event.currentTarget && onClose()}
    >
      <section
        className="dialog"
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="dialog-title"
      >
        <IconButton label="关闭" onClick={onClose}>
          <X size={18} />
        </IconButton>
        <h2 id="dialog-title">{title}</h2>
        <p>{description}</p>
        <footer>
          <Button variant="secondary" onClick={onClose}>
            取消
          </Button>
          <Button
            ref={confirmRef}
            variant={danger ? "danger" : "primary"}
            onClick={onConfirm}
          >
            {confirmLabel}
          </Button>
        </footer>
      </section>
    </div>
  );
}
export function Toast({
  message,
  tone = "success",
  onClose,
}: {
  message: string;
  tone?: "success" | "danger";
  onClose(): void;
}) {
  useEffect(() => {
    const timer = setTimeout(onClose, 4000);
    return () => clearTimeout(timer);
  }, [onClose]);
  return (
    <div className={`toast toast-${tone}`} role="status">
      {message}
      <IconButton label="关闭提示" onClick={onClose}>
        <X size={14} />
      </IconButton>
    </div>
  );
}
