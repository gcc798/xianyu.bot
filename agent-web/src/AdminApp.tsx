import {
  CSSProperties,
  FormEvent,
  ReactNode,
  useCallback,
  useEffect,
  useState,
} from "react";
import AgentApp from "./App";
import { adminApi, CurrentUser, DataRow, PageData } from "./adminApi";
import "./admin.css";

type Field = {
  key: string;
  label: string;
  type?: "text" | "number" | "password" | "boolean" | "select";
  options?: { label: string; value: string | number }[];
  createOnly?: boolean;
  required?: boolean;
};
type Column = {
  key: string;
  label: string;
  render?: (row: DataRow) => ReactNode;
};

const sections = [
  {
    label: "工作台",
    items: [
      ["overview", "总览"],
      ["agent", "智能体"],
    ],
  },
  {
    label: "系统",
    items: [
      ["users", "用户管理"],
      ["dictionaries", "字典管理"],
    ],
  },
  {
    label: "安全",
    items: [["login-logs", "登录日志"]],
  },
] as const;

export default function AdminApp() {
  const [authenticated, setAuthenticated] = useState(
    Boolean(localStorage.getItem("token"))
  );
  const [user, setUser] = useState<CurrentUser | null>(
    () =>
      JSON.parse(
        localStorage.getItem("xy-user") || "null"
      ) as CurrentUser | null
  );
  const [route, setRoute] = useState(
    () => location.hash.replace(/^#\/?/, "") || "overview"
  );

  useEffect(() => {
    const update = () =>
      setRoute(location.hash.replace(/^#\/?/, "") || "overview");
    addEventListener("hashchange", update);
    return () => removeEventListener("hashchange", update);
  }, []);
  useEffect(() => {
    if (!authenticated) return;
    void adminApi
      .currentUser()
      .then(({ userInfo }) => {
        setUser(userInfo);
        localStorage.setItem("xy-user", JSON.stringify(userInfo));
      })
      .catch(() => undefined);
  }, [authenticated]);

  if (!authenticated)
    return (
      <LoginPage
        onLogin={(nextUser) => {
          setAuthenticated(true);
          setUser(nextUser);
        }}
      />
    );

  const logout = async () => {
    try {
      await adminApi.logout();
    } catch {
      /* clear local session even if blacklist is unavailable */
    }
    localStorage.removeItem("token");
    localStorage.removeItem("xy-user");
    setAuthenticated(false);
    setUser(null);
  };
  const isAdmin = Number(user?.authorityId) === 888;
  const visibleSections = isAdmin
    ? sections
    : sections.filter((section) => section.label === "工作台");

  return (
    <div className="admin-app">
      <aside className="admin-sidebar">
        <a className="admin-brand" href="#/overview">
          <span>XY</span>
          <div>
            <strong>Xianyu</strong>
            <small>CONTROL DESK</small>
          </div>
        </a>
        <nav>
          {visibleSections.map((section) => (
            <section key={section.label}>
              <h2>{section.label}</h2>
              {section.items.map(([key, label]) => (
                <a
                  key={key}
                  className={route === key ? "active" : ""}
                  href={`#/${key}`}
                >
                  <span>{label}</span>
                </a>
              ))}
            </section>
          ))}
        </nav>
        <div className="admin-sidebar-foot">
          <span className="live-dot" />
          <div>
            <strong>PostgreSQL</strong>
            <small>GOOSE MANAGED</small>
          </div>
        </div>
      </aside>
      <main className={`admin-main ${route === "agent" ? "agent-route" : ""}`}>
        {route !== "agent" && (
          <header className="admin-topbar">
            <div>
              <span className="crumb">XY /</span> {routeTitle(route)}
            </div>
            <div className="admin-user">
              <span>{initials(user?.nickName || user?.userName || "XY")}</span>
              <div>
                <strong>{user?.nickName || user?.userName || "管理员"}</strong>
                <small>
                  {Number(user?.authorityId) === 888 ? "管理员" : "普通用户"}
                </small>
              </div>
              <button onClick={() => void logout()}>退出</button>
            </div>
          </header>
        )}
        <div className="admin-view">{renderRoute(route, isAdmin)}</div>
      </main>
    </div>
  );
}

function LoginPage({ onLogin }: { onLogin: (user: CurrentUser) => void }) {
  const [form, setForm] = useState({
    username: "admin",
    password: "",
  });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (form.username.length < 5 || form.password.length < 6) {
      setError("用户名至少 5 位，密码至少 6 位");
      return;
    }
    setBusy(true);
    setError("");
    try {
      const data = await adminApi.login(form);
      localStorage.setItem("token", data.token);
      localStorage.setItem("xy-user", JSON.stringify(data.user));
      onLogin(data.user);
    } catch (cause) {
      setError(messageOf(cause));
    } finally {
      setBusy(false);
    }
  };
  return (
    <main className="login-page">
      <section className="login-story">
        <div>
          <h1>
            把复杂业务，
            <br />
            收束在一张桌面。
          </h1>
          <p>账号、自动回复、发货与智能体共享同一套权限边界和审计轨迹。</p>
        </div>
        <footer>
          <i /> Xianyu Automation System
        </footer>
      </section>
      <section className="login-panel">
        <form onSubmit={(event) => void submit(event)}>
          <div className="login-mark">XY</div>
          <h2>欢迎回来</h2>
          <label>
            用户名
            <input
              autoFocus
              autoComplete="username"
              value={form.username}
              onChange={(event) =>
                setForm({ ...form, username: event.target.value })
              }
            />
          </label>
          <label>
            密码
            <input
              type="password"
              autoComplete="current-password"
              value={form.password}
              onChange={(event) =>
                setForm({ ...form, password: event.target.value })
              }
            />
          </label>
          {error && (
            <div className="form-error" role="alert">
              {error}
            </div>
          )}
          <button className="primary wide" disabled={busy}>
            {busy ? "正在验证…" : "进入控制台"}
          </button>
          <small className="login-note">凭据仅发送到当前部署的后端服务</small>
        </form>
      </section>
    </main>
  );
}

function renderRoute(route: string, isAdmin: boolean) {
  if (route === "agent") return <AgentApp />;
  if (route === "overview") return <Overview />;
  if (!isAdmin) return <EmptyState title="仅管理员可访问此页面" />;
  if (route === "users") return <UsersPage />;
  if (route === "dictionaries") return <DictionariesPage />;
  if (route === "login-logs") return <LoginLogsPage />;
  return <EmptyState title="页面不存在" />;
}

function Overview() {
  return (
    <div className="overview-page">
      <div className="overview-hero">
        <div>
          <h1>业务控制台</h1>
          <p>通用后台与智能体已经运行在同一个 React 18 + TypeScript 前端中。</p>
        </div>
        <a href="#/agent">打开智能体</a>
      </div>
      <div className="overview-grid-admin">
        {[
          ["访问边界", "管理员 · 普通用户", "#d8ff5f"],
          ["业务字典", "层级字典与值维护", "#82e7ff"],
          ["登录审计", "成功与失败登录", "#ffc56e"],
          ["数据库", "PostgreSQL 单一基线", "#ff8a72"],
        ].map(([title, text, color]) => (
          <article key={title} style={{ "--accent": color } as CSSProperties}>
            <div>
              <strong>{title}</strong>
              <p>{text}</p>
            </div>
          </article>
        ))}
      </div>
      <section className="migration-note">
        <strong>迁移状态</strong>
        <div>
          <h2>聚焦闲鱼自动化与 Agent</h2>
          <p>
            后台只保留登录、用户、字典与登录审计，业务能力统一由 React
            工程承载。
          </p>
        </div>
      </section>
    </div>
  );
}

function UsersPage() {
  return (
    <ResourcePage
      title="用户管理"
      description="管理后台用户、基础资料与登录状态。"
      rowKey="ID"
      columns={[
        { key: "ID", label: "ID" },
        { key: "userName", label: "用户名" },
        { key: "nickName", label: "昵称" },
        {
          key: "authorityId",
          label: "角色",
          render: (row) =>
            Number(row.authorityId) === 888 ? "管理员" : "普通用户",
        },
        { key: "phone", label: "手机号" },
        { key: "email", label: "邮箱" },
        {
          key: "enable",
          label: "状态",
          render: (row) => (
            <Status
              value={Number(row.enable) === 1 ? "启用" : "停用"}
              good={Number(row.enable) === 1}
            />
          ),
        },
      ]}
      fields={[
        { key: "userName", label: "用户名", required: true },
        {
          key: "password",
          label: "初始密码",
          type: "password",
          createOnly: true,
          required: true,
        },
        { key: "nickName", label: "昵称", required: true },
        {
          key: "authorityId",
          label: "角色",
          type: "select",
          options: [
            { label: "普通用户", value: 9528 },
            { label: "管理员", value: 888 },
          ],
          required: true,
        },
        { key: "phone", label: "手机号" },
        { key: "email", label: "邮箱" },
        {
          key: "enable",
          label: "状态",
          type: "select",
          options: [
            { label: "启用", value: 1 },
            { label: "停用", value: 2 },
          ],
        },
      ]}
      load={(page, search) =>
        adminApi.users({
          page,
          pageSize: 20,
          username: search,
          nickname: search,
          orderKey: "id",
          desc: true,
        })
      }
      create={adminApi.createUser}
      update={adminApi.updateUser}
      remove={(row) => adminApi.deleteUser(row.ID)}
      rowActions={(row) => (
        <button
          onClick={() => {
            const password = window.prompt(
              `为 ${String(row.userName)} 设置新密码（至少 6 位）`
            );
            if (password && password.length >= 6)
              void adminApi.resetPassword(row.ID, password);
          }}
        >
          重置密码
        </button>
      )}
    />
  );
}

function DictionariesPage() {
  const [dictionaries, setDictionaries] = useState<DataRow[]>([]);
  const [selected, setSelected] = useState<DataRow | null>(null);
  const [details, setDetails] = useState<DataRow[]>([]);
  const [editing, setEditing] = useState<{
    kind: "dictionary" | "detail";
    row: DataRow;
  } | null>(null);
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    try {
      const data = await adminApi.dictionaries({ page: 1, pageSize: 1000 });
      const rows = flatten(data);
      setDictionaries(rows);
      setSelected((value) => value || rows[0] || null);
    } catch (cause) {
      setError(messageOf(cause));
    }
  }, []);
  useEffect(() => {
    void load();
  }, [load]);
  useEffect(() => {
    if (!selected) return;
    void adminApi
      .dictionaryDetails(selected.ID)
      .then((data) => setDetails(flatten(data.list)))
      .catch((cause) => setError(messageOf(cause)));
  }, [selected]);
  const save = async (row: DataRow) => {
    if (!editing) return;
    if (editing.kind === "dictionary")
      row.ID
        ? await adminApi.updateDictionary(row)
        : await adminApi.createDictionary(row);
    else
      row.ID
        ? await adminApi.updateDictionaryDetail(row)
        : await adminApi.createDictionaryDetail({
            ...row,
            sysDictionaryID: selected?.ID,
          });
    setEditing(null);
    await load();
    if (selected)
      setDetails(flatten((await adminApi.dictionaryDetails(selected.ID)).list));
  };
  return (
    <Page
      title="字典管理"
      description="保留业务枚举、层级选项和展示标签。"
      actions={
        <button
          className="primary"
          onClick={() =>
            setEditing({ kind: "dictionary", row: { status: true } })
          }
        >
          新建字典
        </button>
      }
    >
      {error && <Notice text={error} />}
      <div className="dictionary-layout">
        <aside className="dictionary-list">
          {dictionaries.map((row) => (
            <button
              key={String(row.ID)}
              className={selected?.ID === row.ID ? "active" : ""}
              onClick={() => setSelected(row)}
            >
              <span>
                {String(row.name)}
                <small>{String(row.type)}</small>
              </span>
              <i>{row.status ? "ON" : "OFF"}</i>
            </button>
          ))}
        </aside>
        <section className="dictionary-detail">
          <header>
            <div>
              <h2>{String(selected?.name || "请选择字典")}</h2>
            </div>
            {selected && (
              <div>
                <button
                  className="secondary"
                  onClick={() =>
                    setEditing({ kind: "dictionary", row: selected })
                  }
                >
                  编辑字典
                </button>
                <button
                  className="primary"
                  onClick={() =>
                    setEditing({
                      kind: "detail",
                      row: {
                        status: true,
                        sort: 0,
                        sysDictionaryID: selected.ID,
                      },
                    })
                  }
                >
                  新增字典项
                </button>
              </div>
            )}
          </header>
          <DataTable
            columns={[
              { key: "label", label: "展示值" },
              { key: "value", label: "字典值" },
              { key: "extend", label: "扩展值" },
              { key: "sort", label: "排序" },
              {
                key: "status",
                label: "状态",
                render: (row) => (
                  <Status
                    value={row.status ? "启用" : "停用"}
                    good={Boolean(row.status)}
                  />
                ),
              },
            ]}
            rows={details}
            actions={(row) => (
              <>
                <button onClick={() => setEditing({ kind: "detail", row })}>
                  编辑
                </button>
                <button
                  className="danger-link"
                  onClick={() =>
                    void confirmDelete("删除该字典项？", () =>
                      adminApi.deleteDictionaryDetail(row.ID)
                    ).then(
                      () =>
                        selected &&
                        adminApi
                          .dictionaryDetails(selected.ID)
                          .then((data) => setDetails(flatten(data.list)))
                    )
                  }
                >
                  删除
                </button>
              </>
            )}
          />
        </section>
      </div>
      {editing && (
        <RecordDialog
          title={editing.row.ID ? "编辑" : "新建"}
          row={editing.row}
          fields={
            editing.kind === "dictionary"
              ? [
                  { key: "name", label: "字典名（中）", required: true },
                  { key: "type", label: "字典类型（英）", required: true },
                  { key: "parentID", label: "父字典 ID", type: "number" },
                  { key: "desc", label: "描述" },
                  { key: "status", label: "状态", type: "boolean" },
                ]
              : [
                  { key: "label", label: "展示值", required: true },
                  { key: "value", label: "字典值", required: true },
                  { key: "extend", label: "扩展值" },
                  { key: "parentID", label: "父项 ID", type: "number" },
                  { key: "sort", label: "排序", type: "number" },
                  { key: "status", label: "状态", type: "boolean" },
                ]
          }
          onClose={() => setEditing(null)}
          onSave={save}
        />
      )}
    </Page>
  );
}

function LoginLogsPage() {
  return (
    <ResourcePage
      title="登录日志"
      description="记录内网后台的成功与失败登录，用于安全追踪。"
      rowKey="ID"
      readOnly
      columns={[
        { key: "ID", label: "ID" },
        { key: "username", label: "用户" },
        { key: "ip", label: "IP" },
        {
          key: "status",
          label: "结果",
          render: (row: DataRow) => (
            <Status
              value={row.status ? "成功" : "失败"}
              good={Boolean(row.status)}
            />
          ),
        },
        { key: "errorMessage", label: "说明" },
        { key: "CreatedAt", label: "时间", render: dateCell },
      ]}
      fields={[]}
      load={(page, search) =>
        adminApi.loginLogs({ page, pageSize: 20, username: search })
      }
      remove={(row) => adminApi.deleteLoginLog(row.ID)}
    />
  );
}

function ResourcePage(props: {
  title: string;
  description: string;
  rowKey: string;
  columns: Column[];
  fields: Field[];
  load: (page: number, search: string) => Promise<PageData>;
  create?: (row: DataRow) => Promise<DataRow>;
  update?: (row: DataRow) => Promise<DataRow>;
  remove?: (row: DataRow) => Promise<DataRow>;
  mapSave?: (row: DataRow) => DataRow;
  extra?: ReactNode;
  rowActions?: (row: DataRow) => ReactNode;
  readOnly?: boolean;
}) {
  const [rows, setRows] = useState<DataRow[]>([]),
    [page, setPage] = useState(1),
    [total, setTotal] = useState(0),
    [search, setSearch] = useState(""),
    [query, setQuery] = useState(""),
    [editing, setEditing] = useState<DataRow | null>(null),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  const load = useCallback(async () => {
    setBusy(true);
    setError("");
    try {
      const data = await props.load(page, query);
      setRows(data.list || []);
      setTotal(data.total || 0);
    } catch (cause) {
      setError(messageOf(cause));
    } finally {
      setBusy(false);
    }
  }, [page, props.load, query]);
  useEffect(() => {
    void load();
  }, [load]);
  const save = async (row: DataRow) => {
    try {
      const data = props.mapSave ? props.mapSave(row) : row;
      if (row[props.rowKey]) await props.update?.(data);
      else await props.create?.(data);
      setEditing(null);
      await load();
    } catch (cause) {
      setError(messageOf(cause));
    }
  };
  return (
    <Page
      title={props.title}
      description={props.description}
      actions={
        <>
          {props.extra}
          {!props.readOnly && (
            <button className="primary" onClick={() => setEditing({})}>
              新建
            </button>
          )}
        </>
      }
    >
      <div className="table-toolbar">
        <form
          onSubmit={(event) => {
            event.preventDefault();
            setPage(1);
            setQuery(search);
          }}
        >
          <input
            aria-label="搜索当前模块"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="搜索当前模块"
          />
          <button>查询</button>
        </form>
        <small>{busy ? "正在加载…" : `${total} 条记录`}</small>
      </div>
      {error && <Notice text={error} />}
      <DataTable
        rows={rows}
        columns={props.columns}
        actions={(row) => (
          <>
            {!props.readOnly && (
              <button onClick={() => setEditing(row)}>编辑</button>
            )}
            {props.rowActions?.(row)}
            {props.remove && (
              <button
                className="danger-link"
                onClick={() =>
                  void confirmDelete("确认删除该记录？", () =>
                    props.remove!(row)
                  ).then(load)
                }
              >
                删除
              </button>
            )}
          </>
        )}
      />
      <Pagination page={page} total={total} onPage={setPage} />
      {editing && (
        <RecordDialog
          title={
            editing[props.rowKey]
              ? `编辑 ${props.title}`
              : `新建 ${props.title}`
          }
          row={editing}
          fields={props.fields}
          onClose={() => setEditing(null)}
          onSave={save}
        />
      )}
    </Page>
  );
}

function Page({
  title,
  description,
  actions,
  children,
}: {
  title: string;
  description: string;
  actions?: ReactNode;
  children: ReactNode;
}) {
  return (
    <section className="resource-page">
      <header className="page-head">
        <div>
          <h1>{title}</h1>
          <p>{description}</p>
        </div>
        <div>{actions}</div>
      </header>
      {children}
    </section>
  );
}
function DataTable({
  rows,
  columns,
  actions,
}: {
  rows: DataRow[];
  columns: Column[];
  actions?: (row: DataRow) => ReactNode;
}) {
  return (
    <div className="table-shell">
      <table>
        <thead>
          <tr>
            {columns.map((column) => (
              <th key={column.key}>{column.label}</th>
            ))}
            {actions && <th>操作</th>}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, index) => (
            <tr key={String(row.ID ?? row.id ?? row.authorityId ?? index)}>
              {columns.map((column) => (
                <td key={column.key}>
                  {column.render ? column.render(row) : cell(row[column.key])}
                </td>
              ))}
              {actions && (
                <td>
                  <div className="row-actions">{actions(row)}</div>
                </td>
              )}
            </tr>
          ))}
          {!rows.length && (
            <tr>
              <td
                className="empty-cell"
                colSpan={columns.length + (actions ? 1 : 0)}
              >
                暂无数据
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

function RecordDialog({
  title,
  row,
  fields,
  onClose,
  onSave,
}: {
  title: string;
  row: DataRow;
  fields: Field[];
  onClose: () => void;
  onSave: (row: DataRow) => Promise<void>;
}) {
  const editing = Boolean(row.ID ?? row.authorityId);
  const visibleFields = fields.filter(
    (field) => !(editing && field.createOnly)
  );
  const [value, setValue] = useState<DataRow>(() => ({ ...row }));
  const [busy, setBusy] = useState(false);
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setBusy(true);
    try {
      await onSave(value);
    } finally {
      setBusy(false);
    }
  };
  return (
    <div
      className="dialog-backdrop"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) onClose();
      }}
    >
      <form
        aria-labelledby="record-dialog-title"
        aria-modal="true"
        className="record-dialog"
        role="dialog"
        onSubmit={(event) => void submit(event)}
      >
        <header>
          <h2 id="record-dialog-title">{title}</h2>
          <button type="button" onClick={onClose}>
            关闭
          </button>
        </header>
        <div className="dialog-fields">
          {visibleFields.map((field) => (
            <label
              key={field.key}
              className={field.type === "boolean" ? "check-field" : ""}
            >
              <span>
                {field.label}
                {field.required && <b>*</b>}
              </span>
              {field.type === "select" ? (
                <select
                  value={String(value[field.key] ?? "")}
                  onChange={(event) =>
                    setValue({
                      ...value,
                      [field.key]: convert(event.target.value, field),
                    })
                  }
                >
                  {field.options?.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              ) : field.type === "boolean" ? (
                <input
                  type="checkbox"
                  checked={Boolean(value[field.key])}
                  onChange={(event) =>
                    setValue({ ...value, [field.key]: event.target.checked })
                  }
                />
              ) : (
                <input
                  required={field.required}
                  type={field.type || "text"}
                  value={String(value[field.key] ?? "")}
                  onChange={(event) =>
                    setValue({
                      ...value,
                      [field.key]: convert(event.target.value, field),
                    })
                  }
                />
              )}
            </label>
          ))}
        </div>
        <footer>
          <button type="button" className="secondary" onClick={onClose}>
            取消
          </button>
          <button className="primary" disabled={busy}>
            {busy ? "保存中…" : "保存更改"}
          </button>
        </footer>
      </form>
    </div>
  );
}

function Pagination({
  page,
  total,
  onPage,
}: {
  page: number;
  total: number;
  onPage: (page: number) => void;
}) {
  const pages = Math.max(1, Math.ceil(total / 20));
  return (
    <div className="pagination">
      <button disabled={page <= 1} onClick={() => onPage(page - 1)}>
        上一页
      </button>
      <span>
        {page} / {pages}
      </span>
      <button disabled={page >= pages} onClick={() => onPage(page + 1)}>
        下一页
      </button>
    </div>
  );
}
function Status({ value, good }: { value: string; good: boolean }) {
  return (
    <span className={`status-pill ${good ? "good" : ""}`}>
      <i />
      {value}
    </span>
  );
}
function Notice({ text }: { text: string }) {
  return (
    <div className="notice" role="alert">
      {text}
    </div>
  );
}
function EmptyState({ title }: { title: string }) {
  return (
    <div className="empty-state">
      <span>404</span>
      <h1>{title}</h1>
      <a href="#/overview">返回总览</a>
    </div>
  );
}
function flatten(rows: DataRow[]): DataRow[] {
  return rows.flatMap((row) => [
    row,
    ...flatten(Array.isArray(row.children) ? (row.children as DataRow[]) : []),
  ]);
}
function messageOf(cause: unknown) {
  return cause instanceof Error ? cause.message : "请求失败";
}
function cell(value: unknown) {
  if (value === null || value === undefined || value === "")
    return <span className="muted">—</span>;
  if (typeof value === "boolean") return value ? "是" : "否";
  if (typeof value === "object")
    return Array.isArray(value) ? `${value.length} 项` : "对象";
  return String(value);
}
function convert(value: string, field: Field) {
  return field.type === "number" ||
    field.options?.some((item) => typeof item.value === "number")
    ? Number(value)
    : value;
}
function dateCell(row: DataRow) {
  const value = row.CreatedAt || row.createdAt;
  return value
    ? new Intl.DateTimeFormat("zh-CN", {
        dateStyle: "short",
        timeStyle: "medium",
      }).format(new Date(String(value)))
    : "—";
}
function initials(value: string) {
  return value.trim().slice(0, 2).toUpperCase();
}
function routeTitle(route: string) {
  for (const section of sections)
    for (const [key, label] of section.items) if (key === route) return label;
  return "控制台";
}
async function confirmDelete(text: string, action: () => Promise<unknown>) {
  if (!window.confirm(text)) return;
  await action();
}
