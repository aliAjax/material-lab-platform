import { useMutation, useQuery } from "@tanstack/react-query";
import {
  Calculator,
  Play,
  RefreshCw,
  Save,
  Send,
  TriangleAlert,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { tasksApi } from "../../api/resources";
import {
  Button,
  ConfirmDialog,
  EmptyState,
  ErrorState,
  Field,
  IconButton,
  LoadingState,
  PageHeader,
  Pagination,
  StatusBadge,
  Toast,
} from "../../components/ui";
import { useUnsavedChanges } from "../../hooks/useUnsavedChanges";

export function TasksPage() {
  const [params, setParams] = useSearchParams();
  const filters = {
    status: params.get("status") || "",
    q: params.get("q") || "",
    cursor: params.get("cursor") || undefined,
    limit: "20",
  };
  const query = useQuery({
    queryKey: ["tasks", filters],
    queryFn: () => tasksApi.list(filters),
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
        title="试验任务"
        description="开始前须确认对应子样已完成交接。"
        actions={
          <IconButton label="刷新任务" onClick={() => void query.refetch()}>
            <RefreshCw size={18} />
          </IconButton>
        }
      />
      <div className="filterbar">
        <input
          placeholder="任务或样品编号"
          value={filters.q}
          onChange={(e) => update("q", e.target.value)}
        />
        <select
          value={filters.status}
          onChange={(e) => update("status", e.target.value)}
        >
          <option value="">全部状态</option>
          <option value="assigned">待开始</option>
          <option value="in_progress">录入中</option>
          <option value="pending_review">待复核</option>
          <option value="completed">已完成</option>
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
                  <th>方法版本</th>
                  <th>执行人</th>
                  <th>轮次</th>
                  <th>状态</th>
                </tr>
              </thead>
              <tbody>
                {query.data!.items.map((item) => (
                  <tr key={item.id}>
                    <td>
                      <Link className="strong-link" to={`/tasks/${item.id}`}>
                        {item.task_no}
                      </Link>
                    </td>
                    <td>{item.sample_no}</td>
                    <td>
                      {item.method_name} v{item.method_version}
                    </td>
                    <td>{item.assignee}</td>
                    <td>{item.round || 1}</td>
                    <td>
                      <StatusBadge value={item.status} tone="info" />
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

export function TaskExecutionPage() {
  const { id } = useParams();
  const query = useQuery({
    queryKey: ["task", id],
    queryFn: () => tasksApi.get(id!),
  });
  const [readings, setReadings] = useState<Record<string, string>>({});
  const [dirty, setDirty] = useState(false);
  const [message, setMessage] = useState("");
  const [submitOpen, setSubmitOpen] = useState(false);
  const [retestOpen, setRetestOpen] = useState(false);
  useUnsavedChanges(dirty);
  useEffect(() => {
    if (query.data?.readings) {
      setReadings(query.data.readings);
      setDirty(false);
    }
  }, [query.data]);
  const preview = useMutation({
    mutationFn: () => tasksApi.preview(id!, readings),
  });
  const save = useMutation({
    mutationFn: () => tasksApi.saveReadings(id!, readings),
    onSuccess: () => {
      setDirty(false);
      setMessage("原始读数已保存");
    },
  });
  const start = useMutation({
    mutationFn: () => tasksApi.start(id!),
    onSuccess: () => void query.refetch(),
  });
  const submit = useMutation({
    mutationFn: () => tasksApi.submit(id!),
    onSuccess: () => {
      setSubmitOpen(false);
      setDirty(false);
      setMessage("已冻结数据并提交复核");
      void query.refetch();
    },
  });
  const retest = useMutation({
    mutationFn: (e: React.FormEvent<HTMLFormElement>) => {
      const data = new FormData(e.currentTarget);
      return tasksApi.retest(
        id!,
        String(data.get("reason")),
        String(data.get("issue_type")),
      );
    },
    onSuccess: () => {
      setRetestOpen(false);
      setMessage("复测申请已提交");
    },
  });
  const canEdit = useMemo(
    () =>
      ["assigned", "in_progress", "returned"].includes(
        query.data?.status || "",
      ),
    [query.data?.status],
  );
  if (query.isPending) return <LoadingState />;
  if (query.isError) return <ErrorState error={query.error} />;
  const task = query.data!;
  return (
    <>
      <PageHeader
        title={task.task_no}
        description={`${task.sample_no} · ${task.method.name} v${task.method.version}`}
        actions={
          task.status === "assigned" ? (
            <Button onClick={() => start.mutate()} icon={<Play size={16} />}>
              确认领样并开始
            </Button>
          ) : (
            <StatusBadge value={task.status} tone="info" />
          )
        }
      />
      <div className="execution-layout">
        <form
          className="form-sheet"
          onSubmit={(e) => {
            e.preventDefault();
            save.mutate();
          }}
        >
          <header>
            <h2>原始读数</h2>
            <span>服务端保留完整精度</span>
          </header>
          <div className="dynamic-fields">
            {task.method.fields.map((field) => (
              <Field
                key={field.key}
                label={`${field.label}${field.unit ? ` (${field.unit})` : ""}`}
                required={field.required}
                hint={
                  field.min || field.max
                    ? `允许范围 ${field.min || "不限"} – ${field.max || "不限"}`
                    : undefined
                }
              >
                {field.type === "boolean" ? (
                  <select
                    disabled={!canEdit}
                    value={readings[field.key] || ""}
                    onChange={(e) => {
                      setReadings({ ...readings, [field.key]: e.target.value });
                      setDirty(true);
                    }}
                  >
                    <option value="">请选择</option>
                    <option value="true">是</option>
                    <option value="false">否</option>
                  </select>
                ) : (
                  <input
                    disabled={!canEdit}
                    inputMode={
                      field.type === "decimal" || field.type === "integer"
                        ? "decimal"
                        : undefined
                    }
                    value={readings[field.key] || ""}
                    onChange={(e) => {
                      setReadings({ ...readings, [field.key]: e.target.value });
                      setDirty(true);
                    }}
                  />
                )}
              </Field>
            ))}
          </div>
          {canEdit && (
            <footer className="form-footer">
              <Button
                variant="secondary"
                type="button"
                onClick={() => preview.mutate()}
                icon={<Calculator size={16} />}
              >
                公式预览
              </Button>
              <Button
                variant="secondary"
                type="submit"
                disabled={!dirty || save.isPending}
                icon={<Save size={16} />}
              >
                保存读数
              </Button>
              <Button
                type="button"
                onClick={() => setSubmitOpen(true)}
                icon={<Send size={16} />}
              >
                提交复核
              </Button>
            </footer>
          )}
        </form>
        <aside className="preview-panel">
          <h2>计算预览</h2>
          <p>此处结果仅供录入核对，最终值由服务端重新计算。</p>
          {preview.isPending ? (
            <LoadingState />
          ) : preview.data ? (
            <dl>
              {Object.entries(preview.data.results).map(([key, value]) => (
                <React.Fragment key={key}>
                  <dt>
                    {task.method.formulas.find((f) => f.key === key)?.label ||
                      key}
                  </dt>
                  <dd>{value}</dd>
                </React.Fragment>
              ))}
            </dl>
          ) : (
            <EmptyState
              title="尚未计算"
              description="填写读数后运行公式预览。"
            />
          )}
          {preview.data?.warnings?.map((w) => (
            <div className="form-alert" key={w}>
              {w}
            </div>
          ))}
        </aside>
      </div>
      <Button
        variant="ghost"
        onClick={() => setRetestOpen(true)}
        icon={<TriangleAlert size={16} />}
      >
        报告异常并申请复测
      </Button>
      {retestOpen && (
        <form
          className="inline-form"
          onSubmit={(e) => {
            e.preventDefault();
            retest.mutate(e);
          }}
        >
          <h2>复测申请</h2>
          <div className="form-grid">
            <Field label="异常类型">
              <select name="issue_type">
                <option value="out_of_range">输入越界</option>
                <option value="equipment">设备异常</option>
                <option value="unstable">结果不稳定</option>
              </select>
            </Field>
            <Field label="申请原因" required>
              <textarea name="reason" required rows={3} />
            </Field>
          </div>
          <footer>
            <Button
              variant="secondary"
              type="button"
              onClick={() => setRetestOpen(false)}
            >
              取消
            </Button>
            <Button type="submit">提交申请</Button>
          </footer>
        </form>
      )}
      <ConfirmDialog
        open={submitOpen}
        title="冻结并提交复核？"
        description="提交后当前修订版的读数、结果和附件不可编辑；退回时将形成新的修订版。"
        confirmLabel="提交复核"
        onClose={() => setSubmitOpen(false)}
        onConfirm={() => submit.mutate()}
      />
      {message && <Toast message={message} onClose={() => setMessage("")} />}
    </>
  );
}

import React from "react";
