import { useMutation, useQuery } from "@tanstack/react-query";
import { Laptop, LogOut } from "lucide-react";
import { authApi } from "../../api/resources";
import {
  Button,
  EmptyState,
  ErrorState,
  LoadingState,
  PageHeader,
  StatusBadge,
} from "../../components/ui";
export function ProfilePage() {
  const query = useQuery({ queryKey: ["sessions"], queryFn: authApi.sessions });
  const revoke = useMutation({
    mutationFn: authApi.revokeSession,
    onSuccess: () => void query.refetch(),
  });
  return (
    <>
      <PageHeader
        title="个人会话"
        description="检查已登录设备并撤销不再使用的刷新会话。"
      />
      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState error={query.error} />
      ) : !query.data!.length ? (
        <EmptyState />
      ) : (
        <div className="session-list">
          {query.data!.map((session) => (
            <article key={session.id}>
              <Laptop size={22} />
              <div>
                <strong>
                  {session.device || "未知设备"}{" "}
                  {session.current && (
                    <StatusBadge value="当前会话" tone="success" />
                  )}
                </strong>
                <span>
                  {session.ip_address} · 最近活动{" "}
                  {new Date(session.last_seen_at).toLocaleString()}
                </span>
              </div>
              {!session.current && (
                <Button
                  variant="danger"
                  disabled={revoke.isPending}
                  onClick={() => revoke.mutate(session.id)}
                  icon={<LogOut size={15} />}
                >
                  撤销
                </Button>
              )}
            </article>
          ))}
        </div>
      )}
    </>
  );
}
