import { useMutation, useQuery } from "@tanstack/react-query";
import { CheckCircle2, GitCompareArrows, RotateCcw } from "lucide-react";
import { useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { reviewsApi } from "../../api/resources";
import {
  Button,
  ConfirmDialog,
  EmptyState,
  ErrorState,
  Field,
  LoadingState,
  PageHeader,
  Pagination,
  StatusBadge,
  Toast,
} from "../../components/ui";

export function ReviewsPage() {
  const [params, setParams] = useSearchParams();
  const filters = {
    status: params.get("status") || "pending",
    cursor: params.get("cursor") || undefined,
    limit: "20",
  };
  const query = useQuery({
    queryKey: ["reviews", filters],
    queryFn: () => reviewsApi.list(filters),
  });
  const update = (key: string, v?: string) => {
    const next = new URLSearchParams(params);
    v ? next.set(key, v) : next.delete(key);
    setParams(next);
  };
  return (
    <>
      <PageHeader
        title="技术复核"
        description="复核员须独立核对原始读数、计算过程、附件与判定。"
      />
      <div className="filterbar">
        <select
          value={filters.status}
          onChange={(e) => update("status", e.target.value)}
        >
          <option value="pending">待复核</option>
          <option value="approved">已通过</option>
          <option value="returned">已退回</option>
        </select>
      </div>
      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState error={query.error} />
      ) : !query.data!.items.length ? (
        <EmptyState />
      ) : (
        <>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>任务</th>
                  <th>样品</th>
                  <th>执行人</th>
                  <th>修订</th>
                  <th>状态</th>
                </tr>
              </thead>
              <tbody>
                {query.data!.items.map((item) => (
                  <tr key={item.id}>
                    <td>
                      <Link className="strong-link" to={`/reviews/${item.id}`}>
                        {item.task_no}
                      </Link>
                    </td>
                    <td>{item.sample_no}</td>
                    <td>{item.executor}</td>
                    <td>v{item.revisions.at(-1)?.version || 1}</td>
                    <td>
                      <StatusBadge value={item.status} tone="warning" />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <Pagination
            previous={query.data!.previous_cursor}
            next={query.data!.next_cursor}
            onChange={(v) => update("cursor", v)}
          />
        </>
      )}
    </>
  );
}
export function ReviewDetailPage() {
  const { id } = useParams();
  const query = useQuery({
    queryKey: ["review", id],
    queryFn: () => reviewsApi.get(id!),
  });
  const [decision, setDecision] = useState<"approve" | "return">();
  const [comment, setComment] = useState("");
  const [message, setMessage] = useState("");
  const mutation = useMutation({
    mutationFn: () => reviewsApi.decide(id!, decision!, comment),
    onSuccess: () => {
      setDecision(undefined);
      setMessage("复核决定已记录");
      void query.refetch();
    },
  });
  if (query.isPending) return <LoadingState />;
  if (query.isError) return <ErrorState error={query.error} />;
  const value = query.data!;
  return (
    <>
      <PageHeader
        title={`复核 ${value.task_no}`}
        description={`${value.sample_no} · 执行人 ${value.executor}`}
        actions={<StatusBadge value={value.status} tone="warning" />}
      />
      <div className="review-grid">
        <section className="detail-section">
          <h2>原始读数</h2>
          <dl>
            {value.readings.map((reading) => (
              <React.Fragment key={reading.key}>
                <dt>{reading.key}</dt>
                <dd>
                  {reading.value} {reading.unit}
                </dd>
              </React.Fragment>
            ))}
          </dl>
        </section>
        <section className="detail-section">
          <h2>服务端计算与判定</h2>
          <dl>
            {value.results.map((result) => (
              <React.Fragment key={result.key}>
                <dt>{result.label}</dt>
                <dd>
                  <strong>
                    {result.value} {result.unit}
                  </strong>
                  {result.decision && (
                    <StatusBadge
                      value={result.decision}
                      tone={result.decision === "PASS" ? "success" : "danger"}
                    />
                  )}
                </dd>
              </React.Fragment>
            ))}
          </dl>
        </section>
      </div>
      <section className="detail-section">
        <h2>
          <GitCompareArrows size={18} />
          版本差异
        </h2>
        {value.revisions.length < 2 ? (
          <EmptyState
            title="首次提交"
            description="当前没有上一修订版可比较。"
          />
        ) : (
          <div className="revision-list">
            {value.revisions.map((revision) => (
              <article key={revision.version}>
                <header>
                  <strong>修订 v{revision.version}</strong>
                  <span>
                    {revision.author} ·{" "}
                    {new Date(revision.created_at).toLocaleString()}
                  </span>
                </header>
                {revision.reason && <p>{revision.reason}</p>}
                {revision.changes && (
                  <table>
                    <thead>
                      <tr>
                        <th>字段</th>
                        <th>修改前</th>
                        <th>修改后</th>
                      </tr>
                    </thead>
                    <tbody>
                      {Object.entries(revision.changes).map(([key, diff]) => (
                        <tr key={key}>
                          <td>{key}</td>
                          <td>
                            <del>{diff.before}</del>
                          </td>
                          <td>
                            <ins>{diff.after}</ins>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </article>
            ))}
          </div>
        )}
      </section>
      {value.status === "pending" && (
        <div className="decision-bar">
          <Field label="复核意见" required={decision === "return"}>
            <textarea
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              rows={2}
            />
          </Field>
          <Button
            variant="danger"
            onClick={() => setDecision("return")}
            icon={<RotateCcw size={16} />}
          >
            退回修订
          </Button>
          <Button
            onClick={() => setDecision("approve")}
            icon={<CheckCircle2 size={16} />}
          >
            通过复核
          </Button>
        </div>
      )}
      <ConfirmDialog
        open={Boolean(decision)}
        title={decision === "approve" ? "确认通过复核？" : "确认退回任务？"}
        description={
          decision === "approve"
            ? "确认本人不是任务执行人，且已逐项核对读数、计算、附件及判定。"
            : "退回后将创建新的可编辑修订版，当前快照会完整保留。"
        }
        danger={decision === "return"}
        confirmLabel={decision === "approve" ? "确认通过" : "确认退回"}
        onClose={() => setDecision(undefined)}
        onConfirm={() => {
          if (decision === "return" && !comment.trim()) {
            setMessage("退回时必须填写复核意见");
            setDecision(undefined);
            return;
          }
          mutation.mutate();
        }}
      />
      {message && (
        <Toast
          message={message}
          tone={message.includes("必须") ? "danger" : "success"}
          onClose={() => setMessage("")}
        />
      )}
    </>
  );
}
import React from "react";
