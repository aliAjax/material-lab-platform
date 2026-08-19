import { zodResolver } from "@hookform/resolvers/zod";
import { FlaskConical, LockKeyhole, UserRound } from "lucide-react";
import { useForm } from "react-hook-form";
import { Navigate, useLocation } from "react-router-dom";
import { z } from "zod";
import { useAuth } from "../../state/AuthContext";
import { Button, Field } from "../../components/ui";

const schema = z.object({
  username: z.string().min(1, "请输入用户名"),
  password: z.string().min(8, "密码至少 8 位"),
});
type LoginValue = z.infer<typeof schema>;
export function LoginPage() {
  const { user, login } = useAuth();
  const location = useLocation();
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginValue>({ resolver: zodResolver(schema) });
  if (user)
    return (
      <Navigate
        to={(location.state as { from?: string })?.from || "/"}
        replace
      />
    );
  return (
    <main className="login-page">
      <section className="login-panel">
        <header>
          <span className="login-logo">
            <FlaskConical size={24} />
          </span>
          <div>
            <h1>材检实验室</h1>
            <p>样品流转与双人复核平台</p>
          </div>
        </header>
        <form
          onSubmit={handleSubmit(async (value) => {
            try {
              await login(value.username, value.password);
            } catch (error) {
              setError("root", {
                message: error instanceof Error ? error.message : "登录失败",
              });
            }
          })}
        >
          <Field label="用户名" error={errors.username?.message} required>
            <div className="input-with-icon">
              <UserRound size={17} />
              <input
                autoFocus
                autoComplete="username"
                {...register("username")}
              />
            </div>
          </Field>
          <Field label="密码" error={errors.password?.message} required>
            <div className="input-with-icon">
              <LockKeyhole size={17} />
              <input
                type="password"
                autoComplete="current-password"
                {...register("password")}
              />
            </div>
          </Field>
          {errors.root && (
            <div className="form-alert" role="alert">
              {errors.root.message}
            </div>
          )}
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "正在验证…" : "登录"}
          </Button>
        </form>
        <footer>连续登录失败将触发临时限制</footer>
      </section>
    </main>
  );
}
