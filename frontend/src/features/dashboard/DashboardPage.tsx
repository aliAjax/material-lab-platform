import { useQuery } from "@tanstack/react-query";
import { ArrowRight, RefreshCw } from "lucide-react";
import { Link } from "react-router-dom";
import { custodyApi, reviewsApi, tasksApi } from "../../api/resources";
import {
  EmptyState,
  ErrorState,
  IconButton,
  LoadingState,
  PageHeader,
  StatusBadge,
} from "../../components/ui";

export function DashboardPage() {
  const tasks = useQuery({
    queryKey: ["tasks", "todo"],
    queryFn: () => tasksApi.list({ status: "actionable", limit: "6" }),
  });
  const custody = useQuery({
    queryKey: ["custody", "pending"],
    queryFn: () =>
      custodyApi.list({ status: "pending_confirmation", limit: "5" }),
  });
  const reviews = useQuery({
    queryKey: ["reviews", "pending"],
    queryFn: () => reviewsApi.list({ status: "pending", limit: "5" }),
  });
  const refresh = () => {
    void tasks.refetch();
    void custody.refetch();
    void reviews.refetch();
  };
  return (
    <>
      <PageHeader
        title="待办工作台"
        description="按职责处理当前需要响应的样品流转事项。"
        actions={
          <IconButton label="刷新待办" onClick={refresh}>
            <RefreshCw size={18} />
          </IconButton>
        }
      />
      <div className="work-columns">
        <TodoSection
          title="我的试验任务"
          href="/tasks"
          query={tasks}
          render={(item: any) => (
            <>
              <div>
                <strong>{item.task_no}</strong>
                <small>
                  {item.sample_no} · {item.method_name} v{item.method_version}
                </small>
              </div>
              <StatusBadge value={item.status} tone="info" />
            </>
          )}
        />
        <TodoSection
          title="待确认交接"
          href="/custody"
          query={custody}
          render={(item: any) => (
            <>
              <div>
                <strong>{item.sample_no}</strong>
                <small>
                  {item.from_user} → {item.to_user}
                </small>
              </div>
              <StatusBadge value="待确认" tone="warning" />
            </>
          )}
        />
        <TodoSection
          title="待技术复核"
          href="/reviews"
          query={reviews}
          render={(item: any) => (
            <>
              <div>
                <strong>{item.task_no}</strong>
                <small>
                  {item.sample_no} · 执行人 {item.executor}
                </small>
              </div>
              <StatusBadge value={item.status} tone="warning" />
            </>
          )}
        />
      </div>
    </>
  );
}
function TodoSection({
  title,
  href,
  query,
  render,
}: {
  title: string;
  href: string;
  query: any;
  render: (item: any) => React.ReactNode;
}) {
  return (
    <section className="work-section">
      <header>
        <h2>{title}</h2>
        <Link to={href}>
          查看全部 <ArrowRight size={15} />
        </Link>
      </header>
      {query.isLoading ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState error={query.error} retry={query.refetch} />
      ) : !query.data?.items.length ? (
        <EmptyState title="已处理完毕" description="当前没有需要处理的事项。" />
      ) : (
        <ul className="todo-list">
          {query.data.items.map((item: any) => (
            <li key={item.id}>
              <Link to={`${href}/${item.id}`}>{render(item)}</Link>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
