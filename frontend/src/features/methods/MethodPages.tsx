import { useMutation, useQuery } from "@tanstack/react-query";
import { CheckCircle2, FilePlus2, Save, Send, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { methodsApi } from "../../api/resources";
import {
  Button,
  ConfirmDialog,
  EmptyState,
  ErrorState,
  Field,
  LoadingState,
  PageHeader,
  StatusBadge,
  Toast,
} from "../../components/ui";
import { useUnsavedChanges } from "../../hooks/useUnsavedChanges";
import type {
  MethodField,
  MethodFormula,
  MethodVersion,
} from "../../types/domain";

export function MethodsPage() {
  const query = useQuery({ queryKey: ["methods"], queryFn: methodsApi.list });
  return (
    <>
      <PageHeader
        title="检测方法版本"
        description="已发布版本不可原地修改；既有任务始终引用创建时的快照。"
        actions={
          <Link className="button button-primary" to="/methods/new">
            <FilePlus2 size={16} />
            新建版本
          </Link>
        }
      />
      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState error={query.error} />
      ) : !query.data!.items.length ? (
        <EmptyState />
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>方法名称</th>
                <th>版本</th>
                <th>适用材料</th>
                <th>字段 / 公式</th>
                <th>状态</th>
                <th>创建时间</th>
              </tr>
            </thead>
            <tbody>
              {query.data!.items.map((item) => (
                <tr key={item.id}>
                  <td>
                    <Link className="strong-link" to={`/methods/${item.id}`}>
                      {item.name}
                    </Link>
                  </td>
                  <td>{item.version}</td>
                  <td>{item.applicable_materials}</td>
                  <td>
                    {item.fields.length} / {item.formulas.length}
                  </td>
                  <td>
                    <StatusBadge
                      value={item.status}
                      tone={item.status === "published" ? "success" : "neutral"}
                    />
                  </td>
                  <td>{new Date(item.created_at).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
}
const empty: Partial<MethodVersion> = {
  name: "",
  version: "1.0.0",
  status: "draft",
  applicable_materials: "",
  decision_rule: "",
  fields: [],
  formulas: [],
};
export function MethodEditorPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const query = useQuery({
    queryKey: ["method", id],
    queryFn: () => methodsApi.get(id!),
    enabled: Boolean(id && id !== "new"),
  });
  const [value, setValue] = useState<Partial<MethodVersion>>(empty);
  const [dirty, setDirty] = useState(false);
  const [message, setMessage] = useState("");
  const [publishOpen, setPublishOpen] = useState(false);
  useUnsavedChanges(dirty);
  useEffect(() => {
    if (query.data) {
      setValue(query.data);
      setDirty(false);
    }
  }, [query.data]);
  const readonly = value.status === "published";
  const update = <K extends keyof MethodVersion>(
    key: K,
    next: MethodVersion[K],
  ) => {
    setValue({ ...value, [key]: next });
    setDirty(true);
  };
  const save = useMutation({
    mutationFn: () => methodsApi.save(value),
    onSuccess: (data) => {
      setValue(data);
      setDirty(false);
      setMessage("方法草稿已保存");
      if (!id || id === "new")
        navigate(`/methods/${data.id}`, { replace: true });
    },
  });
  const validate = useMutation({
    mutationFn: async () => {
      const saved = dirty
        ? await methodsApi.save(value)
        : (value as MethodVersion);
      return methodsApi.validate(saved.id);
    },
    onSuccess: (data) =>
      setMessage(
        data.valid
          ? "方法定义校验通过"
          : `校验失败：${data.errors?.join("；")}`,
      ),
  });
  const publish = useMutation({
    mutationFn: () => methodsApi.publish(value.id!),
    onSuccess: (data) => {
      setValue(data);
      setDirty(false);
      setPublishOpen(false);
      setMessage("方法版本已发布并锁定");
    },
  });
  if (id && id !== "new" && query.isPending) return <LoadingState />;
  if (query.isError) return <ErrorState error={query.error} />;
  return (
    <>
      <PageHeader
        title={
          readonly
            ? `${value.name} v${value.version}`
            : value.id
              ? "编辑方法版本"
              : "新建方法版本"
        }
        description={
          readonly ? "只读历史快照" : "定义输入字段、十进制计算公式与判定规则。"
        }
        actions={
          readonly ? (
            <StatusBadge value="已发布" tone="success" />
          ) : (
            <>
              <Button
                variant="secondary"
                onClick={() => save.mutate()}
                icon={<Save size={16} />}
              >
                保存
              </Button>
              <Button
                variant="secondary"
                onClick={() => validate.mutate()}
                icon={<CheckCircle2 size={16} />}
              >
                校验
              </Button>
              <Button
                onClick={() => setPublishOpen(true)}
                icon={<Send size={16} />}
              >
                发布
              </Button>
            </>
          )
        }
      />
      <div className="form-sheet">
        <div className="form-grid">
          <Field label="方法名称">
            <input
              disabled={readonly}
              value={value.name || ""}
              onChange={(e) => update("name", e.target.value)}
            />
          </Field>
          <Field label="版本号">
            <input
              disabled={readonly}
              value={value.version || ""}
              onChange={(e) => update("version", e.target.value)}
            />
          </Field>
          <Field label="适用材料">
            <input
              disabled={readonly}
              value={value.applicable_materials || ""}
              onChange={(e) => update("applicable_materials", e.target.value)}
            />
          </Field>
          <Field label="判定规则">
            <textarea
              disabled={readonly}
              rows={2}
              value={value.decision_rule || ""}
              onChange={(e) => update("decision_rule", e.target.value)}
            />
          </Field>
        </div>
        <MethodFields
          fields={value.fields || []}
          readonly={readonly}
          onChange={(fields) => update("fields", fields)}
        />
        <MethodFormulas
          formulas={value.formulas || []}
          readonly={readonly}
          onChange={(formulas) => update("formulas", formulas)}
        />
      </div>
      <ConfirmDialog
        open={publishOpen}
        title="发布并锁定版本？"
        description="发布前服务端会检查未知变量、循环依赖、除零风险和非法表达式。发布后任何字段均不可原地修改。"
        confirmLabel="校验并发布"
        onClose={() => setPublishOpen(false)}
        onConfirm={() => publish.mutate()}
      />
      {message && (
        <Toast
          message={message}
          tone={message.includes("失败") ? "danger" : "success"}
          onClose={() => setMessage("")}
        />
      )}
    </>
  );
}
function MethodFields({
  fields,
  readonly,
  onChange,
}: {
  fields: MethodField[];
  readonly: boolean;
  onChange(v: MethodField[]): void;
}) {
  const patch = (i: number, key: keyof MethodField, value: any) =>
    onChange(
      fields.map((f, index) => (index === i ? { ...f, [key]: value } : f)),
    );
  return (
    <section className="editor-section">
      <header>
        <div>
          <h2>输入字段</h2>
          <p>原始读数使用字符串传输，服务端以十进制定点处理。</p>
        </div>
        {!readonly && (
          <Button
            variant="secondary"
            onClick={() =>
              onChange([
                ...fields,
                { key: "", label: "", type: "decimal", required: true },
              ])
            }
          >
            添加字段
          </Button>
        )}
      </header>
      {fields.map((field, i) => (
        <div className="editor-row" key={i}>
          <input
            disabled={readonly}
            aria-label="字段标识"
            placeholder="字段标识"
            value={field.key}
            onChange={(e) => patch(i, "key", e.target.value)}
          />
          <input
            disabled={readonly}
            aria-label="显示名称"
            placeholder="显示名称"
            value={field.label}
            onChange={(e) => patch(i, "label", e.target.value)}
          />
          <select
            disabled={readonly}
            value={field.type}
            onChange={(e) => patch(i, "type", e.target.value)}
          >
            <option value="decimal">小数</option>
            <option value="integer">整数</option>
            <option value="text">文本</option>
            <option value="boolean">布尔</option>
          </select>
          <input
            disabled={readonly}
            aria-label="单位"
            placeholder="单位"
            value={field.unit || ""}
            onChange={(e) => patch(i, "unit", e.target.value)}
          />
          {!readonly && (
            <button
              className="icon-button"
              title="删除字段"
              onClick={() => onChange(fields.filter((_, index) => index !== i))}
            >
              <Trash2 size={16} />
            </button>
          )}
        </div>
      ))}
    </section>
  );
}
function MethodFormulas({
  formulas,
  readonly,
  onChange,
}: {
  formulas: MethodFormula[];
  readonly: boolean;
  onChange(v: MethodFormula[]): void;
}) {
  const patch = (i: number, key: keyof MethodFormula, value: any) =>
    onChange(
      formulas.map((f, index) => (index === i ? { ...f, [key]: value } : f)),
    );
  return (
    <section className="editor-section">
      <header>
        <div>
          <h2>计算公式</h2>
          <p>支持字段引用、四则运算、括号、min、max、avg 与条件比较。</p>
        </div>
        {!readonly && (
          <Button
            variant="secondary"
            onClick={() =>
              onChange([
                ...formulas,
                { key: "", label: "", expression: "", precision: 2 },
              ])
            }
          >
            添加公式
          </Button>
        )}
      </header>
      {formulas.map((formula, i) => (
        <div className="editor-row formula-row" key={i}>
          <input
            disabled={readonly}
            placeholder="结果标识"
            value={formula.key}
            onChange={(e) => patch(i, "key", e.target.value)}
          />
          <input
            disabled={readonly}
            placeholder="显示名称"
            value={formula.label}
            onChange={(e) => patch(i, "label", e.target.value)}
          />
          <input
            disabled={readonly}
            className="formula-input"
            placeholder="avg(a, b)"
            value={formula.expression}
            onChange={(e) => patch(i, "expression", e.target.value)}
          />
          <input
            disabled={readonly}
            type="number"
            min="0"
            max="12"
            aria-label="精度"
            value={formula.precision}
            onChange={(e) => patch(i, "precision", Number(e.target.value))}
          />
          {!readonly && (
            <button
              className="icon-button"
              title="删除公式"
              onClick={() =>
                onChange(formulas.filter((_, index) => index !== i))
              }
            >
              <Trash2 size={16} />
            </button>
          )}
        </div>
      ))}
    </section>
  );
}
