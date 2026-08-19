import { useMutation, useQuery } from "@tanstack/react-query";
import { GitBranch, MapPin, Printer, RefreshCw, Scissors } from "lucide-react";
import { useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { commissionsApi, samplesApi } from "../../api/resources";
import {
  Button,
  EmptyState,
  ErrorState,
  Field,
  IconButton,
  LoadingState,
  PageHeader,
  Pagination,
  StatusBadge,
} from "../../components/ui";
import { NewCommissionShortcut } from "../commissions/CommissionPages";

export function SamplesPage() {
  const [params, setParams] = useSearchParams();
  const filters = {
    q: params.get("q") || "",
    status: params.get("status") || "",
    sort: params.get("sort") || "received_at:desc",
    cursor: params.get("cursor") || undefined,
    limit: "20",
  };
  const query = useQuery({
    queryKey: ["samples", filters],
    queryFn: () => samplesApi.list(filters),
  });
  const update = (key: string, value?: string) => {
    const next = new URLSearchParams(params);
    value ? next.set(key, value) : next.delete(key);
    if (key !== "cursor") next.delete("cursor");
    setParams(next);
  };
  return (
    <>
      <PageHeader
        title="样品与委托"
        description="检索样品身份、流转状态和父子样关系。"
        actions={
          <>
            <IconButton label="刷新列表" onClick={() => void query.refetch()}>
              <RefreshCw size={18} />
            </IconButton>
            <NewCommissionShortcut />
          </>
        }
      />
      <div className="filterbar">
        <input
          aria-label="搜索样品"
          placeholder="样品编号、批号或委托方"
          value={filters.q}
          onChange={(e) => update("q", e.target.value)}
        />
        <select
          aria-label="状态"
          value={filters.status}
          onChange={(e) => update("status", e.target.value)}
        >
          <option value="">全部状态</option>
          <option value="received">已接收</option>
          <option value="split">已拆分</option>
          <option value="testing">检测中</option>
          <option value="sealed">已封存</option>
        </select>
        <select
          aria-label="排序"
          value={filters.sort}
          onChange={(e) => update("sort", e.target.value)}
        >
          <option value="received_at:desc">最近接收</option>
          <option value="sample_no:asc">编号升序</option>
        </select>
      </div>
      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState error={query.error} retry={query.refetch} />
      ) : !query.data!.items.length ? (
        <EmptyState />
      ) : (
        <>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>样品编号</th>
                  <th>标签</th>
                  <th>数量</th>
                  <th>保管位置</th>
                  <th>状态</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {query.data!.items.map((sample) => (
                  <tr key={sample.id}>
                    <td>
                      <Link
                        to={`/samples/${sample.id}`}
                        className="strong-link"
                      >
                        {sample.sample_no}
                      </Link>
                    </td>
                    <td>{sample.label}</td>
                    <td>
                      {sample.quantity} {sample.unit}
                    </td>
                    <td>{sample.location || "未指定"}</td>
                    <td>
                      <StatusBadge value={sample.status} />
                    </td>
                    <td>
                      <Link to={`/samples/${sample.id}`}>查看</Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <Pagination
            previous={query.data!.previous_cursor}
            next={query.data!.next_cursor}
            onChange={(value) => update("cursor", value)}
          />
        </>
      )}
    </>
  );
}
export function SampleDetailPage() {
  const { id } = useParams();
  const [splitOpen, setSplitOpen] = useState(false);
  const sample = useQuery({
    queryKey: ["sample", id],
    queryFn: () => samplesApi.get(id!),
  });
  const commission = useQuery({
    queryKey: ["commission-for-sample", id],
    queryFn: () => commissionsApi.get(id!),
    retry: false,
  });
  const split = useMutation({
    mutationFn: (event: React.FormEvent<HTMLFormElement>) => {
      const data = Object.fromEntries(new FormData(event.currentTarget));
      return samplesApi.split(id!, data);
    },
    onSuccess: () => {
      setSplitOpen(false);
      void sample.refetch();
    },
  });
  if (sample.isPending) return <LoadingState />;
  if (sample.isError) return <ErrorState error={sample.error} />;
  const value = sample.data!;
  return (
    <>
      <PageHeader
        title={value.sample_no}
        description={`${value.label} · ${value.quantity} ${value.unit}`}
        actions={
          <>
            <Link
              className="button button-secondary"
              to={`/samples/${id}/label`}
            >
              <Printer size={16} />
              标签
            </Link>
            <Button
              onClick={() => setSplitOpen(!splitOpen)}
              icon={<Scissors size={16} />}
            >
              拆分子样
            </Button>
          </>
        }
      />
      <div className="detail-grid">
        <section className="detail-section">
          <h2>样品身份</h2>
          <dl>
            <dt>状态</dt>
            <dd>
              <StatusBadge value={value.status} tone="info" />
            </dd>
            <dt>保管位置</dt>
            <dd>
              <MapPin size={15} />
              {value.location || "未指定"}
            </dd>
            <dt>用途</dt>
            <dd>{value.purpose || "原始样品"}</dd>
            <dt>损耗说明</dt>
            <dd>{value.loss_reason || "无"}</dd>
          </dl>
        </section>
        <section className="detail-section">
          <h2>委托来源</h2>
          {commission.data ? (
            <dl>
              <dt>委托方</dt>
              <dd>{commission.data.client_name}</dd>
              <dt>材料牌号</dt>
              <dd>{commission.data.material_grade}</dd>
              <dt>批号</dt>
              <dd>{commission.data.batch_no}</dd>
              <dt>检测要求</dt>
              <dd>{commission.data.test_requirements}</dd>
            </dl>
          ) : (
            <p className="muted">委托来源信息暂不可用</p>
          )}
        </section>
      </div>
      {splitOpen && (
        <form
          className="inline-form"
          onSubmit={(e) => {
            e.preventDefault();
            split.mutate(e);
          }}
        >
          <h2>拆分新子样</h2>
          <div className="form-grid">
            <Field label="用途" required>
              <input name="purpose" required />
            </Field>
            <Field label="标签" required>
              <input name="label" required />
            </Field>
            <Field label="数量" required>
              <input name="quantity" inputMode="decimal" required />
            </Field>
            <Field label="单位" required>
              <input name="unit" defaultValue={value.unit} required />
            </Field>
            <Field label="保管位置">
              <input name="location" />
            </Field>
            <Field label="损耗原因" hint="父样与子样数量无法守恒时必填">
              <input name="loss_reason" />
            </Field>
          </div>
          <footer>
            <Button
              variant="secondary"
              type="button"
              onClick={() => setSplitOpen(false)}
            >
              取消
            </Button>
            <Button type="submit" disabled={split.isPending}>
              确认拆分
            </Button>
          </footer>
        </form>
      )}
      <section className="detail-section tree-section">
        <h2>
          <GitBranch size={18} />
          父子样关系
        </h2>
        {!value.children?.length ? (
          <EmptyState title="尚未拆分" description="该样品当前没有子样。" />
        ) : (
          <div className="sample-tree">
            <strong>{value.sample_no}</strong>
            {value.children.map((child) => (
              <Link key={child.id} to={`/samples/${child.id}`}>
                <span>{child.sample_no}</span>
                <small>
                  {child.purpose} · {child.quantity} {child.unit}
                </small>
              </Link>
            ))}
          </div>
        )}
      </section>
    </>
  );
}
export function SampleLabelPage() {
  const { id } = useParams();
  const query = useQuery({
    queryKey: ["sample", id],
    queryFn: () => samplesApi.get(id!),
  });
  if (query.isPending) return <LoadingState />;
  if (query.isError) return <ErrorState error={query.error} />;
  return (
    <main className="label-page">
      <div className="label-ticket">
        <span>材检实验室样品</span>
        <h1>{query.data.sample_no}</h1>
        <dl>
          <dt>标签</dt>
          <dd>{query.data.label}</dd>
          <dt>数量</dt>
          <dd>
            {query.data.quantity} {query.data.unit}
          </dd>
          <dt>位置</dt>
          <dd>{query.data.location || "-"}</dd>
        </dl>
      </div>
      <Button
        className="no-print"
        onClick={() => window.print()}
        icon={<Printer size={16} />}
      >
        打印标签
      </Button>
    </main>
  );
}
