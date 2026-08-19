import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowRightLeft, Check, RotateCcw } from "lucide-react";
import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import { custodyApi } from "../../api/resources";
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

export function CustodyPage() {
  const [params, setParams] = useSearchParams();
  const client = useQueryClient();
  const [creating, setCreating] = useState(false);
  const [confirmId, setConfirmId] = useState<string>();
  const [message, setMessage] = useState("");
  const filters = {
    status: params.get("status") || "pending_confirmation",
    kind: params.get("kind") || "",
    cursor: params.get("cursor") || undefined,
    limit: "20",
  };
  const query = useQuery({
    queryKey: ["custody", filters],
    queryFn: () => custodyApi.list(filters),
  });
  const refresh = () => client.invalidateQueries({ queryKey: ["custody"] });
  const confirm = useMutation({
    mutationFn: (id: string) => custodyApi.confirm(id),
    onSuccess: () => {
      setConfirmId(undefined);
      setMessage("交接已由双方确认");
      void refresh();
    },
  });
  const create = useMutation({
    mutationFn: (event: React.FormEvent<HTMLFormElement>) =>
      custodyApi.create(Object.fromEntries(new FormData(event.currentTarget))),
    onSuccess: () => {
      setCreating(false);
      setMessage("已发起交接，等待对方确认");
      void refresh();
    },
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
        title="样品交接链"
        description="领取、转交、归还和封存须由交出人与接收人分别确认。"
        actions={
          <Button
            onClick={() => setCreating(!creating)}
            icon={<ArrowRightLeft size={16} />}
          >
            发起交接
          </Button>
        }
      />
      {creating && (
        <form
          className="inline-form"
          onSubmit={(e) => {
            e.preventDefault();
            create.mutate(e);
          }}
        >
          <h2>新交接记录</h2>
          <div className="form-grid">
            <Field label="样品编号" required>
              <input name="sample_no" required />
            </Field>
            <Field label="类型">
              <select name="kind">
                <option value="checkout">领取</option>
                <option value="transfer">转交</option>
                <option value="return">归还</option>
                <option value="seal">封存</option>
              </select>
            </Field>
            <Field label="接收人账号" required>
              <input name="to_user_id" required />
            </Field>
            <Field label="目的位置" required>
              <input name="to_location" required />
            </Field>
            <Field label="样品状态" required>
              <input name="condition" required />
            </Field>
            <Field label="备注">
              <input name="note" />
            </Field>
          </div>
          <footer>
            <Button
              type="button"
              variant="secondary"
              onClick={() => setCreating(false)}
            >
              取消
            </Button>
            <Button type="submit" disabled={create.isPending}>
              提交交接
            </Button>
          </footer>
        </form>
      )}
      <div className="filterbar">
        <select
          value={filters.status}
          onChange={(e) => update("status", e.target.value)}
        >
          <option value="">全部状态</option>
          <option value="pending_confirmation">待确认</option>
          <option value="confirmed">已确认</option>
          <option value="reversed">已冲正</option>
        </select>
        <select
          value={filters.kind}
          onChange={(e) => update("kind", e.target.value)}
        >
          <option value="">全部类型</option>
          <option value="checkout">领取</option>
          <option value="transfer">转交</option>
          <option value="return">归还</option>
          <option value="seal">封存</option>
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
                  <th>样品</th>
                  <th>动作</th>
                  <th>交出人 → 接收人</th>
                  <th>位置变化</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {query.data!.items.map((item) => (
                  <tr key={item.id}>
                    <td>{item.sample_no}</td>
                    <td>{item.kind}</td>
                    <td>
                      {item.from_user} → {item.to_user}
                    </td>
                    <td>
                      {item.from_location} → {item.to_location}
                    </td>
                    <td>
                      <StatusBadge
                        value={item.status}
                        tone={
                          item.status === "pending_confirmation"
                            ? "warning"
                            : "success"
                        }
                      />
                    </td>
                    <td className="row-actions">
                      {item.status === "pending_confirmation" && (
                        <Button
                          variant="secondary"
                          onClick={() => setConfirmId(item.id)}
                          icon={<Check size={15} />}
                        >
                          确认
                        </Button>
                      )}
                      {item.status === "confirmed" && (
                        <ReverseButton
                          id={item.id}
                          onDone={() => void refresh()}
                        />
                      )}
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
      <ConfirmDialog
        open={Boolean(confirmId)}
        title="确认已接收样品？"
        description="确认后交接记录不可修改。请先核对样品标签、状态与所在位置。"
        confirmLabel="确认接收"
        onClose={() => setConfirmId(undefined)}
        onConfirm={() => confirmId && confirm.mutate(confirmId)}
      />
      {message && <Toast message={message} onClose={() => setMessage("")} />}
    </>
  );
}
function ReverseButton({ id, onDone }: { id: string; onDone(): void }) {
  const [open, setOpen] = useState(false);
  const mutation = useMutation({
    mutationFn: () =>
      custodyApi.reverse(
        id,
        (document.querySelector(`#reason-${id}`) as HTMLInputElement)?.value ||
          "",
      ),
    onSuccess: () => {
      setOpen(false);
      onDone();
    },
  });
  return (
    <>
      {
        <Button
          variant="ghost"
          onClick={() => setOpen(true)}
          icon={<RotateCcw size={15} />}
        >
          冲正
        </Button>
      }
      <ConfirmDialog
        open={open}
        title="新增冲正记录？"
        description="不会修改原记录，将新增一条关联的反向审计记录。确认前请确保已记录具体原因。"
        confirmLabel="继续冲正"
        danger
        onClose={() => setOpen(false)}
        onConfirm={() => mutation.mutate()}
      />
      {open && (
        <input
          id={`reason-${id}`}
          className="visually-hidden"
          defaultValue="交接记录纠错"
        />
      )}
    </>
  );
}
