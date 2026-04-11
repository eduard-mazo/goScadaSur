import { useState, useEffect, useMemo, useCallback } from "react";
import {
  LayoutDashboard,
  Search,
  Database,
  FileCode,
  Settings,
  Activity,
  AlertTriangle,
  ChevronRight,
  Save,
  RefreshCw,
  Server,
  ShieldCheck,
  UserCircle,
  User as UserIcon,
  LogOut,
  Play,
  Trash,
  Terminal,
  Plus,
  Trash2,
  Code,
  FileText,
  Download,
  Eye,
  X,
  Info,
  AlertCircle,
  CheckCircle2,
} from "lucide-react";
import "./App.css";
import {
  getAppConfig,
  saveAppConfig,
  getTemplateStats,
  getDasipConfig,
  saveDasipConfig,
  getRawTemplates,
  saveRawTemplates,
  uploadFile,
  login,
  logout,
  runQuery,
  searchStation,
  listOutputFiles,
  getOutputFile,
  type AppConfig,
  type TemplateStats,
  type DasipConfig,
  type SearchStationRequest,
  type UploadResponse,
  type OutputFileItem,
} from "./api";

type View = "dashboard" | "search" | "query" | "generator" | "settings";

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isVisitor, setIsVisitor] = useState(false);
  const [userData, setUserData] = useState<{
    username: string;
    role: string;
  } | null>(null);
  const [currentView, setCurrentView] = useState<View>("dashboard");
  const [appConfig, setAppConfig] = useState<AppConfig | null>(null);
  const [templateStats, setTemplateStats] = useState<TemplateStats | null>(
    null,
  );
  const [loading, setLoading] = useState(true);
  const [isConnected, setIsConnected] = useState(false);
  const [isDbActive, setIsDbActive] = useState(false);

  const fetchData = useCallback(async () => {
    if (!isAuthenticated && !isVisitor) return;
    try {
      const configData = await getAppConfig();
      const stats = await getTemplateStats();
      setAppConfig(configData);
      setTemplateStats(stats);
      setIsConnected(true);
      setIsDbActive(
        !!configData.postgres?.host && configData.postgres.port > 0,
      );
    } catch (error: unknown) {
      console.error("Error al actualizar datos:", error);
      setIsConnected(false);
      // Si recibimos 401 y no somos visitante, forzar logout
      if (typeof error === "object" && error !== null && "response" in error) {
        const axiosError = error as { response: { status: number } };
        if (axiosError.response.status === 401 && !isVisitor) {
          setIsAuthenticated(false);
          localStorage.removeItem("gs_token");
        }
      }
    }
  }, [isAuthenticated, isVisitor]);

  const checkInitialState = useCallback(async () => {
    try {
      const configData = await getAppConfig();
      const dbActive =
        !!configData.postgres?.host && configData.postgres.port > 0;
      setIsDbActive(dbActive);
      setAppConfig(configData);
      setIsConnected(true);

      const savedUser = localStorage.getItem("gs_user");
      const token = localStorage.getItem("gs_token");
      const savedVisitor = localStorage.getItem("gs_is_visitor") === "true";

      if (savedUser && token) {
        setUserData(JSON.parse(savedUser));
        setIsAuthenticated(true);
        setIsVisitor(false);
      } else if (savedVisitor) {
        setIsVisitor(true);
        setIsAuthenticated(true);
      } else {
        setIsAuthenticated(false);
      }

      if (token || savedVisitor) {
        const stats = await getTemplateStats();
        setTemplateStats(stats);
      }
    } catch (error: unknown) {
      console.error("Error en verificación inicial:", error);
      setIsConnected(false);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    checkInitialState();
  }, [checkInitialState]);

  const handleLogout = () => {
    localStorage.removeItem("gs_is_visitor");
    logout();
  };

  useEffect(() => {
    if (isAuthenticated) {
      const interval = setInterval(fetchData, 30000);
      return () => clearInterval(interval);
    }
  }, [isAuthenticated, fetchData]);

  if (loading) {
    return (
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          height: "100vh",
          gap: "1.5rem",
          background: "#f3f4f6",
        }}
      >
        <Activity size={60} className="spinning" color="#38b449" />
        <p style={{ fontWeight: 800, color: "#111827", fontSize: "1.2rem" }}>
          Iniciando Motor goScadaSur...
        </p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <LoginView
        onLoginSuccess={() => {
          setIsAuthenticated(true);
          checkInitialState();
        }}
        onVisitorEntry={() => {
          setIsVisitor(true);
          setIsAuthenticated(true);
          localStorage.setItem("gs_is_visitor", "true");
          checkInitialState();
        }}
        isDbActive={isDbActive}
      />
    );
  }

  const renderView = () => {
    switch (currentView) {
      case "dashboard":
        return (
          <DashboardView
            stats={templateStats}
            onRefresh={fetchData}
            isDbActive={isDbActive}
          />
        );
      case "search":
        return <SearchView />;
      case "query":
        return <QueryView />;
      case "generator":
        return <GeneratorView />;
      case "settings":
        return <SettingsView config={appConfig} onUpdate={fetchData} />;
      default:
        return (
          <DashboardView
            stats={templateStats}
            onRefresh={fetchData}
            isDbActive={isDbActive}
          />
        );
    }
  };

  return (
    <div className="app-container">
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="sidebar-logo">GS</div>
          <h1 className="sidebar-title">goScadaSur</h1>
        </div>

        <nav className="sidebar-nav">
          <button
            className={`nav-item ${currentView === "dashboard" ? "active" : ""}`}
            onClick={() => setCurrentView("dashboard")}
          >
            <LayoutDashboard size={22} />
            <span>Tablero Principal</span>
          </button>
          <button
            className={`nav-item ${currentView === "search" ? "active" : ""}`}
            onClick={() => setCurrentView("search")}
          >
            <Search size={22} />
            <span>Búsqueda Estación</span>
          </button>
          <button
            className={`nav-item ${currentView === "query" ? "active" : ""}`}
            onClick={() => setCurrentView("query")}
          >
            <Database size={22} />
            <span>Consulta SQL</span>
          </button>
          <button
            className={`nav-item ${currentView === "generator" ? "active" : ""}`}
            onClick={() => setCurrentView("generator")}
          >
            <FileCode size={22} />
            <span>Generador XML</span>
          </button>
        </nav>

        <div className="sidebar-footer">
          <button
            className={`nav-item ${currentView === "settings" ? "active" : ""}`}
            onClick={() => setCurrentView("settings")}
          >
            <Settings size={22} />
            <span>Configuración</span>
          </button>

          <div className="user-profile">
            <div className="avatar">
              {isVisitor ? <UserCircle size={20} /> : <UserIcon size={20} />}
            </div>
            <div className="user-details">
              <span className="name">
                {userData?.username ||
                  (isVisitor ? "Visitante Externo" : "Usuario")}
              </span>
              <span className="role">
                {isVisitor
                  ? "Acceso Limitado"
                  : userData?.role === "admin"
                    ? "Administrador"
                    : "Operador"}
              </span>
            </div>
          </div>

          <button className="btn-outline-danger" onClick={handleLogout}>
            <LogOut size={18} /> <span>Cerrar Sesión</span>
          </button>
        </div>
      </aside>

      <main className="main-content">
        <header className="header">
          <div style={{ display: "flex", alignItems: "center", gap: "0.6rem" }}>
            <span
              style={{
                color: "var(--text-secondary)",
                fontSize: "0.95rem",
                fontWeight: 600,
              }}
            >
              Sistema
            </span>
            <ChevronRight size={16} color="#9ca3af" />
            <span
              style={{
                fontWeight: 800,
                color: "var(--epm-dark-green)",
                fontSize: "1rem",
              }}
            >
              {currentView === "dashboard"
                ? "Tablero Principal"
                : currentView === "search"
                  ? "Búsqueda de Estaciones"
                  : currentView === "query"
                    ? "Motor de Consultas"
                    : currentView === "generator"
                      ? "Procesador de Datos"
                      : "Configuración"}
            </span>
          </div>

          <div className="status-container">
            <div
              className={`status-pill ${isConnected ? "online" : "offline"}`}
              title="Estado del Servidor API"
            >
              <div className="status-dot"></div>
              <Server size={16} />
              <span>Servidor API</span>
            </div>
            <div
              className={`status-pill ${isDbActive ? "online" : "offline"}`}
              title="Estado de PostgreSQL"
            >
              <div className="status-dot"></div>
              <Database size={16} />
              <span>{isDbActive ? "Persistencia" : "Sin Base Datos"}</span>
            </div>
          </div>
        </header>

        <div className="view-body">{renderView()}</div>
      </main>
    </div>
  );
}

const DashboardView = ({
  stats,
  onRefresh,
  isDbActive,
}: {
  stats: TemplateStats | null;
  onRefresh: () => void;
  isDbActive: boolean;
}) => (
  <div>
    <div
      style={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        marginBottom: "2.5rem",
      }}
    >
      <div>
        <h2 style={{ fontSize: "1.75rem", fontWeight: 800, color: "#111827" }}>
          Estado del Sistema
        </h2>
        <p style={{ color: "var(--text-secondary)", fontWeight: 600 }}>
          Estadísticas en tiempo real de la configuración actual
        </p>
      </div>
      <button className="btn-epm" onClick={onRefresh}>
        <RefreshCw size={18} /> Actualizar Datos
      </button>
    </div>

    {!isDbActive && (
      <div
        style={{
          background: "#fffbeb",
          border: "1px solid #f59e0b",
          padding: "1.25rem",
          borderRadius: "var(--radius-md)",
          display: "flex",
          alignItems: "center",
          gap: "1rem",
          marginBottom: "2.5rem",
          color: "#92400e",
        }}
      >
        <AlertTriangle size={24} />
        <div style={{ fontWeight: 700 }}>
          Modo Limitado: Conecte PostgreSQL en Configuración para habilitar el
          historial de trabajos y flujo de workflow.
        </div>
      </div>
    )}

    <div
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(4, 1fr)",
        gap: "1.5rem",
      }}
    >
      <div className="stat-box">
        <div className="stat-value">{stats?.stats.total || 0}</div>
        <div className="stat-label">Plantillas Totales</div>
      </div>
      <div className="stat-box">
        <div className="stat-value">{stats?.stats.analog || 0}</div>
        <div className="stat-label">Señales Análogas</div>
      </div>
      <div className="stat-box">
        <div className="stat-value">{stats?.stats.discrete || 0}</div>
        <div className="stat-label">Señales Digitales</div>
      </div>
      <div className="stat-box">
        <div className="stat-value">{stats?.stats.breaker || 0}</div>
        <div className="stat-label">Interruptores</div>
      </div>
    </div>
  </div>
);

const SettingsView = ({
  config,
  onUpdate,
}: {
  config: AppConfig | null;
  onUpdate: () => void;
}) => {
  const [activeTab, setActiveTab] = useState<"general" | "dasip" | "templates">(
    "general",
  );
  const [localConfig, setLocalConfig] = useState<AppConfig | null>(null);
  const [dasip, setDasip] = useState<DasipConfig | null>(null);
  const [rawTemplates, setRawTemplates] = useState<string>("");
  const [templateEditorMode, setTemplateEditorMode] = useState<"form" | "json">(
    "form",
  );
  const [selectedTemplateKey, setSelectedTemplateKey] = useState<string | null>(
    null,
  );
  const [searchTerm, setSearchTerm] = useState("");
  const [saving, setSaving] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (config) setLocalConfig(JSON.parse(JSON.stringify(config)));
  }, [config]);

  useEffect(() => {
    const loadTabData = async () => {
      if (activeTab === "general") return;
      setLoading(true);
      try {
        if (activeTab === "dasip") {
          const data = await getDasipConfig();
          setDasip(data);
        } else if (activeTab === "templates") {
          const data = await getRawTemplates();
          setRawTemplates(data);
        }
      } catch {
        console.error("Error cargando datos:");
      } finally {
        setLoading(false);
      }
    };
    loadTabData();
  }, [activeTab]);

  const parsedTemplates = useMemo(() => {
    try {
      return JSON.parse(rawTemplates);
    } catch {
      return {};
    }
  }, [rawTemplates]);

  const filteredTemplateKeys = useMemo(() => {
    return Object.keys(parsedTemplates)
      .filter((key) => key.toLowerCase().includes(searchTerm.toLowerCase()))
      .sort();
  }, [parsedTemplates, searchTerm]);

  const handleSaveGeneral = async () => {
    if (!localConfig) return;
    setSaving(true);
    try {
      await saveAppConfig(localConfig);
      alert("Configuración guardada correctamente");
      onUpdate();
    } catch {
      alert("Error guardando configuración");
    } finally {
      setSaving(false);
    }
  };

  const handleSaveDasip = async () => {
    if (!dasip) return;
    setSaving(true);
    try {
      await saveDasipConfig(dasip);
      alert("Mapeo de Red DASIP aplicado");
      onUpdate();
    } catch {
      alert("Error guardando DASIP");
    } finally {
      setSaving(false);
    }
  };

  const handleSaveTemplates = async (content: string) => {
    setSaving(true);
    try {
      await saveRawTemplates(content);
      setRawTemplates(content);
      alert("Plantillas XML actualizadas");
      onUpdate();
    } catch {
      alert("Error guardando plantillas. Verifique el formato JSON.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="dashboard-card">
      <div style={{ marginBottom: "2.5rem" }}>
        <h2
          style={{
            fontSize: "1.5rem",
            fontWeight: 800,
            marginBottom: "1rem",
            display: "flex",
            alignItems: "center",
            gap: "0.75rem",
          }}
        >
          <Settings size={24} color="var(--epm-green)" /> Configuración del
          Sistema
        </h2>
        <div
          style={{
            display: "flex",
            gap: "1rem",
            borderBottom: "1px solid var(--border)",
            paddingBottom: "0.5rem",
          }}
        >
          <button
            className={`nav-item ${activeTab === "general" ? "active" : ""}`}
            onClick={() => setActiveTab("general")}
            style={{ width: "auto", padding: "0.5rem 1rem" }}
          >
            Motor de App
          </button>
          <button
            className={`nav-item ${activeTab === "dasip" ? "active" : ""}`}
            onClick={() => setActiveTab("dasip")}
            style={{ width: "auto", padding: "0.5rem 1rem" }}
          >
            Mapa de Red
          </button>
          <button
            className={`nav-item ${activeTab === "templates" ? "active" : ""}`}
            onClick={() => setActiveTab("templates")}
            style={{ width: "auto", padding: "0.5rem 1rem" }}
          >
            Plantillas XML
          </button>
        </div>
      </div>

      {loading ? (
        <div style={{ padding: "4rem", textAlign: "center" }}>
          <Activity size={40} className="spinning" color="var(--epm-green)" />
        </div>
      ) : (
        <>
          {activeTab === "general" && localConfig && (
            <div
              style={{ display: "flex", flexDirection: "column", gap: "2rem" }}
            >
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "1.5rem",
                }}
              >
                <div
                  className="stat-box"
                  style={{ borderLeftColor: "#3b82f6" }}
                >
                  <h3
                    style={{
                      fontSize: "1rem",
                      fontWeight: 800,
                      marginBottom: "1.25rem",
                    }}
                  >
                    Contexto de Aplicación
                  </h3>
                  <div
                    style={{
                      display: "flex",
                      flexDirection: "column",
                      gap: "1rem",
                    }}
                  >
                    <div className="form-group">
                      <label
                        className="form-label"
                        style={{ fontWeight: 700, fontSize: "0.85rem" }}
                      >
                        Nombre del Sistema
                      </label>
                      <input
                        type="text"
                        value={localConfig.app.name}
                        onChange={(e) =>
                          setLocalConfig({
                            ...localConfig,
                            app: { ...localConfig.app, name: e.target.value },
                          })
                        }
                        style={{
                          width: "100%",
                          padding: "0.5rem",
                          borderRadius: "4px",
                          border: "1px solid #ccc",
                        }}
                      />
                    </div>
                    <div className="form-group">
                      <label
                        className="form-label"
                        style={{ fontWeight: 700, fontSize: "0.85rem" }}
                      >
                        Descripción
                      </label>
                      <textarea
                        value={localConfig.app.description}
                        onChange={(e) =>
                          setLocalConfig({
                            ...localConfig,
                            app: {
                              ...localConfig.app,
                              description: e.target.value,
                            },
                          })
                        }
                        style={{
                          width: "100%",
                          height: "80px",
                          padding: "0.5rem",
                          borderRadius: "4px",
                          border: "1px solid #ccc",
                        }}
                      />
                    </div>
                  </div>
                </div>

                <div
                  className="stat-box"
                  style={{ borderLeftColor: "#8b5cf6" }}
                >
                  <h3
                    style={{
                      fontSize: "1rem",
                      fontWeight: 800,
                      marginBottom: "1.25rem",
                    }}
                  >
                    Capa de Persistencia (PostgreSQL)
                  </h3>
                  <div
                    style={{
                      display: "grid",
                      gridTemplateColumns: "1fr 1fr",
                      gap: "1rem",
                    }}
                  >
                    <div className="form-group">
                      <label
                        className="form-label"
                        style={{ fontWeight: 700, fontSize: "0.85rem" }}
                      >
                        Servidor (Host)
                      </label>
                      <input
                        type="text"
                        value={localConfig.postgres.host}
                        onChange={(e) =>
                          setLocalConfig({
                            ...localConfig,
                            postgres: {
                              ...localConfig.postgres,
                              host: e.target.value,
                            },
                          })
                        }
                        style={{
                          width: "100%",
                          padding: "0.5rem",
                          borderRadius: "4px",
                          border: "1px solid #ccc",
                        }}
                      />
                    </div>
                    <div className="form-group">
                      <label
                        className="form-label"
                        style={{ fontWeight: 700, fontSize: "0.85rem" }}
                      >
                        Puerto
                      </label>
                      <input
                        type="number"
                        value={localConfig.postgres.port}
                        onChange={(e) =>
                          setLocalConfig({
                            ...localConfig,
                            postgres: {
                              ...localConfig.postgres,
                              port: parseInt(e.target.value),
                            },
                          })
                        }
                        style={{
                          width: "100%",
                          padding: "0.5rem",
                          borderRadius: "4px",
                          border: "1px solid #ccc",
                        }}
                      />
                    </div>
                    <div className="form-group">
                      <label
                        className="form-label"
                        style={{ fontWeight: 700, fontSize: "0.85rem" }}
                      >
                        Usuario
                      </label>
                      <input
                        type="text"
                        value={localConfig.postgres.user}
                        onChange={(e) =>
                          setLocalConfig({
                            ...localConfig,
                            postgres: {
                              ...localConfig.postgres,
                              user: e.target.value,
                            },
                          })
                        }
                        style={{
                          width: "100%",
                          padding: "0.5rem",
                          borderRadius: "4px",
                          border: "1px solid #ccc",
                        }}
                      />
                    </div>
                    <div className="form-group">
                      <label
                        className="form-label"
                        style={{ fontWeight: 700, fontSize: "0.85rem" }}
                      >
                        Base de Datos
                      </label>
                      <input
                        type="text"
                        value={localConfig.postgres.dbname}
                        onChange={(e) =>
                          setLocalConfig({
                            ...localConfig,
                            postgres: {
                              ...localConfig.postgres,
                              dbname: e.target.value,
                            },
                          })
                        }
                        style={{
                          width: "100%",
                          padding: "0.5rem",
                          borderRadius: "4px",
                          border: "1px solid #ccc",
                        }}
                      />
                    </div>
                  </div>
                </div>
              </div>

              <button
                className="btn-epm"
                onClick={handleSaveGeneral}
                disabled={saving}
                style={{ alignSelf: "flex-start" }}
              >
                <Save size={18} />{" "}
                {saving ? "Guardando..." : "Aplicar Configuración Global"}
              </button>
            </div>
          )}

          {activeTab === "dasip" && (
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "1.5rem",
              }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <h3 style={{ fontSize: "1.25rem", fontWeight: 800 }}>
                  Infraestructura de Red (DASIP)
                </h3>
                <div style={{ display: "flex", gap: "1rem" }}>
                  <button
                    className="btn-epm"
                    onClick={() => {
                      const key = prompt("Nueva IP/Key DASIP:");
                      if (key)
                        setDasip((d) =>
                          d
                            ? {
                                ...d,
                                dasip_mapping: {
                                  ...d.dasip_mapping,
                                  [key]: "",
                                },
                              }
                            : null,
                        );
                    }}
                  >
                    <Plus size={18} /> Agregar Entrada
                  </button>
                  <button
                    className="btn-epm"
                    onClick={handleSaveDasip}
                    disabled={saving}
                    style={{ background: "var(--epm-dark-green)" }}
                  >
                    <Save size={18} />{" "}
                    {saving ? "Guardando..." : "Guardar Mapeo"}
                  </button>
                </div>
              </div>

              <div
                className="form-group"
                style={{
                  background: "#f8fafc",
                  padding: "1rem",
                  borderRadius: "8px",
                }}
              >
                <label className="form-label" style={{ fontWeight: 700 }}>
                  Ruta Global por Defecto
                </label>
                <input
                  type="text"
                  value={dasip?.default_path || ""}
                  onChange={(e) =>
                    setDasip((d) =>
                      d ? { ...d, default_path: e.target.value } : null,
                    )
                  }
                  style={{
                    width: "100%",
                    padding: "0.6rem",
                    border: "1px solid #d1d5db",
                    borderRadius: "4px",
                  }}
                />
              </div>

              <div
                style={{
                  border: "1px solid #e5e7eb",
                  borderRadius: "8px",
                  overflow: "hidden",
                }}
              >
                <table
                  style={{
                    width: "100%",
                    borderCollapse: "collapse",
                    fontSize: "0.9rem",
                  }}
                >
                  <thead style={{ background: "#f9fafb" }}>
                    <tr>
                      <th
                        style={{
                          padding: "1rem",
                          textAlign: "left",
                          borderBottom: "1px solid #e5e7eb",
                        }}
                      >
                        Identificador DASIP (IP)
                      </th>
                      <th
                        style={{
                          padding: "1rem",
                          textAlign: "left",
                          borderBottom: "1px solid #e5e7eb",
                        }}
                      >
                        Ruta de Red Survalent
                      </th>
                      <th
                        style={{
                          padding: "1rem",
                          textAlign: "center",
                          borderBottom: "1px solid #e5e7eb",
                          width: "80px",
                        }}
                      >
                        Acción
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {Object.entries(dasip?.dasip_mapping || {}).map(
                      ([key, val]) => (
                        <tr key={key}>
                          <td
                            style={{
                              padding: "0.75rem 1rem",
                              borderBottom: "1px solid #f3f4f6",
                              fontWeight: 700,
                              color: "var(--epm-dark-green)",
                            }}
                          >
                            {key}
                          </td>
                          <td
                            style={{
                              padding: "0.75rem 1rem",
                              borderBottom: "1px solid #f3f4f6",
                            }}
                          >
                            <input
                              type="text"
                              value={val}
                              onChange={(e) => {
                                const newMapping = {
                                  ...dasip!.dasip_mapping,
                                  [key]: e.target.value,
                                };
                                setDasip({
                                  ...dasip!,
                                  dasip_mapping: newMapping,
                                });
                              }}
                              style={{
                                width: "100%",
                                padding: "0.4rem",
                                border: "1px solid transparent",
                                background: "transparent",
                              }}
                              onFocus={(e) => {
                                (
                                  e.target as HTMLInputElement
                                ).style.borderColor = "#d1d5db";
                              }}
                              onBlur={(e) => {
                                (
                                  e.target as HTMLInputElement
                                ).style.borderColor = "transparent";
                              }}
                            />
                          </td>
                          <td
                            style={{
                              padding: "0.75rem 1rem",
                              borderBottom: "1px solid #f3f4f6",
                              textAlign: "center",
                            }}
                          >
                            <button
                              onClick={() => {
                                if (!dasip) return;
                                const newMapping = { ...dasip.dasip_mapping };
                                delete newMapping[key];
                                setDasip({
                                  ...dasip,
                                  dasip_mapping: newMapping,
                                });
                              }}
                              style={{
                                background: "transparent",
                                border: "none",
                                cursor: "pointer",
                                color: "#ef4444",
                              }}
                            >
                              <Trash2 size={18} />
                            </button>
                          </td>
                        </tr>
                      ),
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {activeTab === "templates" && (
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "1.5rem",
              }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <h3 style={{ fontSize: "1.25rem", fontWeight: 800 }}>
                  Gestor de Plantillas XML
                </h3>
                <div style={{ display: "flex", gap: "1rem" }}>
                  <div
                    style={{
                      display: "flex",
                      background: "#f3f4f6",
                      padding: "0.25rem",
                      borderRadius: "8px",
                    }}
                  >
                    <button
                      onClick={() => setTemplateEditorMode("form")}
                      style={{
                        padding: "0.4rem 0.8rem",
                        border: "none",
                        borderRadius: "6px",
                        background:
                          templateEditorMode === "form"
                            ? "white"
                            : "transparent",
                        cursor: "pointer",
                        fontWeight: 600,
                      }}
                    >
                      Formulario
                    </button>
                    <button
                      onClick={() => setTemplateEditorMode("json")}
                      style={{
                        padding: "0.4rem 0.8rem",
                        border: "none",
                        borderRadius: "6px",
                        background:
                          templateEditorMode === "json"
                            ? "white"
                            : "transparent",
                        cursor: "pointer",
                        fontWeight: 600,
                      }}
                    >
                      JSON
                    </button>
                  </div>
                  <button
                    className="btn-epm"
                    onClick={() => handleSaveTemplates(rawTemplates)}
                    disabled={saving}
                  >
                    <Save size={18} /> Guardar Cambios
                  </button>
                </div>
              </div>

              {templateEditorMode === "json" ? (
                <textarea
                  value={rawTemplates}
                  onChange={(e) => setRawTemplates(e.target.value)}
                  style={{
                    width: "100%",
                    height: "500px",
                    background: "#1e1e1e",
                    color: "#d4d4d4",
                    fontFamily: "monospace",
                    padding: "1.5rem",
                    borderRadius: "8px",
                    fontSize: "0.95rem",
                  }}
                />
              ) : (
                <div
                  style={{
                    gridTemplateColumns: "300px 1fr",
                    display: "grid",
                    gap: "2rem",
                    height: "600px",
                  }}
                >
                  <div
                    style={{
                      background: "#f8fafc",
                      borderRadius: "8px",
                      padding: "1rem",
                      overflowY: "auto",
                    }}
                  >
                    <input
                      type="text"
                      placeholder="Buscar plantilla..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      style={{
                        width: "100%",
                        padding: "0.5rem",
                        marginBottom: "1rem",
                        borderRadius: "4px",
                        border: "1px solid #d1d5db",
                      }}
                    />
                    {filteredTemplateKeys.map((key: string) => (
                      <div
                        key={key}
                        onClick={() => setSelectedTemplateKey(key)}
                        style={{
                          padding: "0.75rem",
                          borderRadius: "6px",
                          cursor: "pointer",
                          background:
                            selectedTemplateKey === key
                              ? "var(--epm-dark-green)"
                              : "transparent",
                          color:
                            selectedTemplateKey === key ? "white" : "inherit",
                          fontWeight: 600,
                          fontSize: "0.85rem",
                          marginBottom: "0.25rem",
                        }}
                      >
                        {key}
                      </div>
                    ))}
                  </div>
                  <div
                    style={{
                      background: "white",
                      border: "1px dashed #d1d5db",
                      borderRadius: "8px",
                      padding: "2rem",
                    }}
                  >
                    {selectedTemplateKey ? (
                      <div>
                        <h4
                          style={{
                            fontSize: "1.2rem",
                            color: "var(--epm-dark-green)",
                            marginBottom: "1.5rem",
                          }}
                        >
                          {selectedTemplateKey}
                        </h4>
                        <pre
                          style={{
                            background: "#f1f5f9",
                            padding: "1rem",
                            borderRadius: "4px",
                            fontSize: "0.8rem",
                          }}
                        >
                          {JSON.stringify(
                            parsedTemplates[selectedTemplateKey],
                            null,
                            2,
                          )}
                        </pre>
                        <p
                          style={{
                            marginTop: "1rem",
                            fontSize: "0.85rem",
                            color: "#64748b",
                          }}
                        >
                          Utilice el modo JSON para ediciones avanzadas de la
                          estructura XML.
                        </p>
                      </div>
                    ) : (
                      <div
                        style={{
                          textAlign: "center",
                          marginTop: "5rem",
                          color: "#94a3b8",
                        }}
                      >
                        <Code
                          size={48}
                          style={{ marginBottom: "1rem", opacity: 0.3 }}
                        />
                        <p>
                          Seleccione una plantilla para editar sus propiedades
                        </p>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}
        </>
      )}
    </div>
  );
};

const SearchView = () => {
  const [form, setForm] = useState<SearchStationRequest>({
    host: "",
    path: "",
    user: "",
    password: "",
    aor: "1",
  });
  const [results, setResults] = useState<{
    b3: string;
    empresa: string;
    region: string;
    result: unknown;
  } | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSearch = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await searchStation(form);
      setResults(
        res as { b3: string; empresa: string; region: string; result: unknown },
      );
    } catch (error: unknown) {
      if (typeof error === "object" && error !== null && "response" in error) {
        const axiosError = error as { response: { data: { error: string } } };
        setError(
          axiosError.response.data.error || "Error al ejecutar la búsqueda",
        );
      } else {
        setError("Error al ejecutar la búsqueda");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="dashboard-card">
      <div style={{ marginBottom: "2rem" }}>
        <h2
          style={{
            fontSize: "1.5rem",
            fontWeight: 800,
            display: "flex",
            alignItems: "center",
            gap: "0.75rem",
          }}
        >
          <Search size={24} color="var(--epm-green)" /> Búsqueda de Estaciones
          Survalent
        </h2>
        <p style={{ color: "var(--text-secondary)", fontWeight: 600 }}>
          Extraer configuración de señales directamente desde la base de datos
          SCADA
        </p>
      </div>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: "1.5rem",
          marginBottom: "1.5rem",
        }}
      >
        <div className="form-group">
          <label className="form-label" style={{ fontWeight: 700 }}>
            Host de Base de Datos
          </label>
          <input
            type="text"
            className="search-input"
            style={{
              width: "100%",
              padding: "0.75rem",
              borderRadius: "8px",
              border: "1px solid var(--border)",
            }}
            placeholder="e.g. 10.0.0.1"
            value={form.host}
            onChange={(e) => setForm({ ...form, host: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label className="form-label" style={{ fontWeight: 700 }}>
            Ruta del Sistema (Path)
          </label>
          <input
            type="text"
            className="search-input"
            style={{
              width: "100%",
              padding: "0.75rem",
              borderRadius: "8px",
              border: "1px solid var(--border)",
            }}
            placeholder="EMPRESA/REGION/B1/B2/B3"
            value={form.path}
            onChange={(e) => setForm({ ...form, path: e.target.value })}
          />
        </div>
      </div>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr 1fr",
          gap: "1.5rem",
          marginBottom: "2rem",
        }}
      >
        <div className="form-group">
          <label className="form-label" style={{ fontWeight: 700 }}>
            Usuario DB
          </label>
          <input
            type="text"
            className="search-input"
            style={{
              width: "100%",
              padding: "0.75rem",
              borderRadius: "8px",
              border: "1px solid var(--border)",
            }}
            value={form.user}
            onChange={(e) => setForm({ ...form, user: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label className="form-label" style={{ fontWeight: 700 }}>
            Contraseña DB
          </label>
          <input
            type="password"
            className="search-input"
            style={{
              width: "100%",
              padding: "0.75rem",
              borderRadius: "8px",
              border: "1px solid var(--border)",
            }}
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label className="form-label" style={{ fontWeight: 700 }}>
            AOR (Área de Responsabilidad)
          </label>
          <input
            type="text"
            className="search-input"
            style={{
              width: "100%",
              padding: "0.75rem",
              borderRadius: "8px",
              border: "1px solid var(--border)",
            }}
            value={form.aor}
            onChange={(e) => setForm({ ...form, aor: e.target.value })}
          />
        </div>
      </div>

      <button
        className="btn-epm"
        onClick={handleSearch}
        disabled={loading}
        style={{ marginBottom: "2rem" }}
      >
        {loading ? (
          <Activity size={20} className="spinning" />
        ) : (
          <Search size={20} />
        )}
        Iniciar Operación de Búsqueda
      </button>

      {error && (
        <div
          style={{
            background: "#fef2f2",
            color: "#991b1b",
            padding: "1rem",
            borderRadius: "8px",
            marginBottom: "2rem",
            borderLeft: "4px solid #ef4444",
            fontWeight: 700,
          }}
        >
          {error}
        </div>
      )}

      {results && (
        <div
          className="stat-box"
          style={{ borderLeftColor: "var(--epm-green)", background: "#fff" }}
        >
          <h3
            style={{
              fontSize: "1.1rem",
              fontWeight: 800,
              marginBottom: "1rem",
            }}
          >
            Resultados de la Estación: {results.b3}
          </h3>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "2rem",
            }}
          >
            <div>
              <p style={{ fontSize: "0.9rem", color: "var(--text-secondary)" }}>
                <strong>Empresa:</strong> {results.empresa}
              </p>
              <p style={{ fontSize: "0.9rem", color: "var(--text-secondary)" }}>
                <strong>Región:</strong> {results.region}
              </p>
            </div>
            <div style={{ textAlign: "right" }}>
              <div className="status-pill online">
                Datos Cargados en Memoria
              </div>
            </div>
          </div>
          <div
            style={{
              marginTop: "1.5rem",
              maxHeight: "300px",
              overflow: "auto",
              background: "#f8fafc",
              padding: "1rem",
              borderRadius: "8px",
              fontSize: "0.85rem",
            }}
          >
            <pre>{JSON.stringify(results.result, null, 2)}</pre>
          </div>
        </div>
      )}
    </div>
  );
};

const QueryView = () => {
  const [query, setQuery] = useState("SELECT * FROM Signals LIMIT 10;");
  const [results, setResults] = useState<Record<string, unknown>[] | null>(
    null,
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleExecute = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await runQuery({ query });
      setResults(res as Record<string, unknown>[]);
    } catch (error: unknown) {
      if (typeof error === "object" && error !== null && "response" in error) {
        const axiosError = error as { response: { data: { error: string } } };
        setError(
          axiosError.response.data.error || "Error al ejecutar la consulta",
        );
      } else {
        setError("Error al ejecutar la consulta");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="dashboard-card"
      style={{
        height: "calc(100vh - 12rem)",
        display: "flex",
        flexDirection: "column",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "1.5rem",
        }}
      >
        <h2
          style={{
            fontSize: "1.5rem",
            fontWeight: 800,
            display: "flex",
            alignItems: "center",
            gap: "0.75rem",
          }}
        >
          <Terminal size={24} color="var(--epm-green)" /> Consola SQL Avanzada
        </h2>
        <div style={{ display: "flex", gap: "1rem" }}>
          <button
            className="btn-epm"
            onClick={handleExecute}
            disabled={loading}
          >
            {loading ? (
              <Activity size={18} className="spinning" />
            ) : (
              <Play size={18} />
            )}
            Ejecutar (F5)
          </button>
          <button
            className="btn-outline-danger"
            style={{ width: "auto" }}
            onClick={() => setResults(null)}
          >
            <Trash size={18} /> Limpiar
          </button>
        </div>
      </div>

      <div className="sql-console">
        <textarea
          className="sql-textarea"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          spellCheck={false}
        />

        {error && (
          <div
            style={{
              color: "#ef4444",
              padding: "1rem",
              background: "rgba(239, 68, 68, 0.1)",
              borderRadius: "4px",
              fontSize: "0.9rem",
              borderLeft: "4px solid #ef4444",
            }}
          >
            <strong>Error SQL:</strong> {error}
          </div>
        )}

        <div className="sql-results">
          {results ? (
            <div style={{ overflow: "auto" }}>
              <table style={{ width: "100%", borderCollapse: "collapse" }}>
                <thead>
                  <tr>
                    {results.length > 0 &&
                      Object.keys(results[0]).map((key) => (
                        <th
                          key={key}
                          style={{
                            textAlign: "left",
                            padding: "0.5rem",
                            borderBottom: "1px solid #444",
                            color: "#fff",
                          }}
                        >
                          {key}
                        </th>
                      ))}
                  </tr>
                </thead>
                <tbody>
                  {results.map((row, i) => (
                    <tr key={i}>
                      {Object.values(row).map((val: unknown, j) => (
                        <td
                          key={j}
                          style={{
                            padding: "0.5rem",
                            borderBottom: "1px solid #333",
                          }}
                        >
                          {String(val)}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div
              style={{ color: "#666", textAlign: "center", marginTop: "2rem" }}
            >
              Esperando ejecución... Los resultados aparecerán aquí.
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

const GeneratorView = () => {
  const [isDragging, setIsDragging] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [result, setResult] = useState<UploadResponse | null>(null);
  const [outputs, setOutputs] = useState<OutputFileItem[]>([]);
  const [loadingList, setLoadingList] = useState(false);
  const [logFilter, setLogFilter] = useState<"all" | "info" | "warn" | "error">(
    "all",
  );
  const [previewName, setPreviewName] = useState<string | null>(null);
  const [previewContent, setPreviewContent] = useState<string>("");

  const refreshList = useCallback(async () => {
    setLoadingList(true);
    try {
      const data = await listOutputFiles();
      setOutputs(data);
    } catch (e) {
      console.error("Error listando outputs:", e);
    } finally {
      setLoadingList(false);
    }
  }, []);

  useEffect(() => {
    refreshList();
  }, [refreshList]);

  const processFile = async (file: File | undefined) => {
    if (!file) return;
    setResult(null);
    setUploading(true);
    try {
      const res = await uploadFile(file);
      setResult(res);
      await refreshList();
    } catch (error: unknown) {
      const fallback: UploadResponse = {
        file: file.name,
        success: false,
        logs: [],
        summary: { info: 0, warn: 0, error: 1 },
        files_generated: [],
        error: "Error procesando archivo",
      };
      if (typeof error === "object" && error !== null && "response" in error) {
        const ax = error as { response: { data: UploadResponse } };
        setResult(ax.response.data ?? fallback);
      } else {
        setResult(fallback);
      }
    } finally {
      setUploading(false);
      setIsDragging(false);
    }
  };

  const openPreview = async (name: string) => {
    setPreviewName(name);
    setPreviewContent("Cargando...");
    try {
      const text = await getOutputFile(name);
      setPreviewContent(text);
    } catch {
      setPreviewContent("Error cargando archivo");
    }
  };

  const filteredLogs = (result?.logs ?? []).filter(
    (l) => logFilter === "all" || l.level === logFilter,
  );
  const levelColor = (lvl: string) =>
    lvl === "error" ? "#ef4444" : lvl === "warn" ? "#f59e0b" : "#0ea5e9";
  const levelIcon = (lvl: string) =>
    lvl === "error" ? (
      <AlertCircle size={14} />
    ) : lvl === "warn" ? (
      <AlertTriangle size={14} />
    ) : (
      <Info size={14} />
    );

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1.5rem" }}>
      <div className="dashboard-card">
        <h2
          style={{
            fontSize: "1.5rem",
            fontWeight: 800,
            marginBottom: "1rem",
            display: "flex",
            alignItems: "center",
            gap: "0.75rem",
          }}
        >
          <FileCode size={24} color="var(--epm-green)" /> Generador XML masivo
          (IFS/IMM)
        </h2>
        <p style={{ color: "var(--text-secondary)", marginBottom: "1.5rem" }}>
          Sube un CSV/XLSX con muchas estaciones. Se agruparán por{" "}
          <code>EMPRESA/REGION/B1/B2/B3</code> y se generará un par IFS+IMM por
          cada una.
        </p>

        <div
          onDragOver={(e) => {
            e.preventDefault();
            setIsDragging(true);
          }}
          onDragLeave={() => setIsDragging(false)}
          onDrop={(e) => {
            e.preventDefault();
            processFile(e.dataTransfer.files[0]);
          }}
          onClick={() =>
            !uploading && document.getElementById("file-upload")?.click()
          }
          style={{
            border: `3px dashed ${isDragging ? "#38b449" : "#d1d5db"}`,
            padding: "3rem",
            borderRadius: "12px",
            textAlign: "center",
            cursor: uploading ? "wait" : "pointer",
            background: isDragging ? "#f0fdf4" : "#f9fafb",
            transition: "all 0.2s",
          }}
        >
          {uploading ? (
            <>
              <Activity
                size={56}
                className="spinning"
                color="#38b449"
                style={{ marginBottom: "1rem" }}
              />
              <p style={{ fontWeight: 800, fontSize: "1.1rem" }}>
                Procesando archivo...
              </p>
            </>
          ) : (
            <>
              <FileCode
                size={56}
                color={isDragging ? "#38b449" : "#9ca3af"}
                style={{ marginBottom: "1rem" }}
              />
              <p style={{ fontWeight: 800, fontSize: "1.2rem" }}>
                Arrastra tu archivo aquí o haz click para seleccionar
              </p>
              <p style={{ color: "#6b7280", marginTop: "0.5rem" }}>
                Formatos: .csv, .xlsx, .xls
              </p>
            </>
          )}
          <input
            type="file"
            id="file-upload"
            style={{ display: "none" }}
            accept=".csv,.xlsx,.xls"
            onChange={(e) => processFile(e.target.files?.[0])}
          />
        </div>
      </div>

      {result && (
        <div className="dashboard-card">
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: "1rem",
            }}
          >
            <h3
              style={{
                fontSize: "1.15rem",
                fontWeight: 800,
                display: "flex",
                alignItems: "center",
                gap: "0.5rem",
              }}
            >
              {result.success ? (
                <CheckCircle2 size={22} color="#16a34a" />
              ) : (
                <AlertCircle size={22} color="#ef4444" />
              )}
              Resultado: {result.file}
            </h3>
            <div style={{ display: "flex", gap: "0.5rem" }}>
              <span
                className="status-pill online"
                style={{
                  background: "#eff6ff",
                  borderColor: "#0ea5e9",
                  color: "#075985",
                }}
              >
                <Info size={14} /> {result.summary.info} info
              </span>
              <span className="status-pill offline">
                <AlertTriangle size={14} /> {result.summary.warn} warn
              </span>
              <span
                className="status-pill"
                style={{
                  background: "#fef2f2",
                  border: "1px solid #ef4444",
                  color: "#991b1b",
                }}
              >
                <AlertCircle size={14} /> {result.summary.error} error
              </span>
            </div>
          </div>

          {result.error && (
            <div
              style={{
                background: "#fef2f2",
                color: "#991b1b",
                padding: "0.75rem 1rem",
                borderRadius: "8px",
                marginBottom: "1rem",
                borderLeft: "4px solid #ef4444",
                fontWeight: 700,
                fontSize: "0.9rem",
              }}
            >
              {result.error}
            </div>
          )}

          {result.files_generated.length > 0 && (
            <div style={{ marginBottom: "1.25rem" }}>
              <p
                style={{
                  fontWeight: 700,
                  fontSize: "0.85rem",
                  color: "var(--text-secondary)",
                  marginBottom: "0.5rem",
                }}
              >
                {result.files_generated.length} archivos generados:
              </p>
              <div style={{ display: "flex", flexWrap: "wrap", gap: "0.5rem" }}>
                {result.files_generated.map((f) => (
                  <button
                    key={f}
                    onClick={() => openPreview(f)}
                    style={{
                      background: "#f0fdf4",
                      border: "1px solid #86efac",
                      color: "#166534",
                      padding: "0.4rem 0.75rem",
                      borderRadius: "6px",
                      fontSize: "0.8rem",
                      fontWeight: 700,
                      cursor: "pointer",
                      display: "flex",
                      alignItems: "center",
                      gap: "0.4rem",
                    }}
                  >
                    <FileText size={14} /> {f}
                  </button>
                ))}
              </div>
            </div>
          )}

          <div
            style={{ display: "flex", gap: "0.5rem", marginBottom: "0.75rem" }}
          >
            {(["all", "info", "warn", "error"] as const).map((lvl) => (
              <button
                key={lvl}
                onClick={() => setLogFilter(lvl)}
                style={{
                  padding: "0.4rem 0.85rem",
                  borderRadius: "6px",
                  fontSize: "0.8rem",
                  fontWeight: 700,
                  border: "1px solid #d1d5db",
                  cursor: "pointer",
                  background:
                    logFilter === lvl ? "var(--epm-dark-green)" : "white",
                  color: logFilter === lvl ? "white" : "#374151",
                }}
              >
                {lvl.toUpperCase()}
              </button>
            ))}
          </div>

          <div
            style={{
              background: "#0f172a",
              borderRadius: "8px",
              padding: "1rem",
              maxHeight: "320px",
              overflow: "auto",
              fontFamily: "monospace",
              fontSize: "0.8rem",
            }}
          >
            {filteredLogs.length === 0 ? (
              <p style={{ color: "#64748b" }}>(sin entradas)</p>
            ) : (
              filteredLogs.map((l, i) => (
                <div
                  key={i}
                  style={{
                    display: "flex",
                    gap: "0.6rem",
                    padding: "0.25rem 0",
                    color: "#cbd5e1",
                  }}
                >
                  <span style={{ color: "#64748b", minWidth: "70px" }}>
                    {new Date(l.time).toLocaleTimeString()}
                  </span>
                  <span
                    style={{
                      color: levelColor(l.level),
                      display: "inline-flex",
                      alignItems: "center",
                      gap: "0.25rem",
                      minWidth: "60px",
                    }}
                  >
                    {levelIcon(l.level)} {l.level}
                  </span>
                  <span style={{ flex: 1, wordBreak: "break-word" }}>
                    {l.msg}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>
      )}

      <div className="dashboard-card">
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "1rem",
          }}
        >
          <h3 style={{ fontSize: "1.15rem", fontWeight: 800 }}>
            Archivos en output/
          </h3>
          <button
            className="btn-epm"
            onClick={refreshList}
            disabled={loadingList}
            style={{ padding: "0.5rem 1rem", fontSize: "0.85rem" }}
          >
            <RefreshCw size={16} className={loadingList ? "spinning" : ""} />{" "}
            Refrescar
          </button>
        </div>
        {outputs.length === 0 ? (
          <p style={{ color: "#94a3b8", textAlign: "center", padding: "2rem" }}>
            No hay XMLs generados todavía.
          </p>
        ) : (
          <div
            style={{
              border: "1px solid #e5e7eb",
              borderRadius: "8px",
              overflow: "hidden",
            }}
          >
            <table
              style={{
                width: "100%",
                borderCollapse: "collapse",
                fontSize: "0.88rem",
              }}
            >
              <thead style={{ background: "#f9fafb" }}>
                <tr>
                  <th style={{ padding: "0.75rem 1rem", textAlign: "left" }}>
                    Archivo
                  </th>
                  <th style={{ padding: "0.75rem 1rem", textAlign: "right" }}>
                    Tamaño
                  </th>
                  <th style={{ padding: "0.75rem 1rem", textAlign: "left" }}>
                    Modificado
                  </th>
                  <th
                    style={{
                      padding: "0.75rem 1rem",
                      textAlign: "center",
                      width: "160px",
                    }}
                  >
                    Acción
                  </th>
                </tr>
              </thead>
              <tbody>
                {outputs.map((f) => (
                  <tr key={f.name} style={{ borderTop: "1px solid #f3f4f6" }}>
                    <td
                      style={{
                        padding: "0.6rem 1rem",
                        fontWeight: 600,
                        color: "var(--epm-dark-green)",
                      }}
                    >
                      <FileText
                        size={14}
                        style={{ display: "inline", marginRight: "0.4rem" }}
                      />
                      {f.name}
                    </td>
                    <td
                      style={{
                        padding: "0.6rem 1rem",
                        textAlign: "right",
                        color: "#64748b",
                      }}
                    >
                      {(f.size / 1024).toFixed(1)} KB
                    </td>
                    <td style={{ padding: "0.6rem 1rem", color: "#64748b" }}>
                      {new Date(f.modified).toLocaleString()}
                    </td>
                    <td style={{ padding: "0.6rem 1rem", textAlign: "center" }}>
                      <button
                        onClick={() => openPreview(f.name)}
                        style={{
                          background: "transparent",
                          border: "1px solid #d1d5db",
                          borderRadius: "6px",
                          padding: "0.3rem 0.6rem",
                          cursor: "pointer",
                          marginRight: "0.4rem",
                        }}
                      >
                        <Eye size={14} />
                      </button>
                      <a
                        href={`http://localhost:8080/api/output/file?name=${encodeURIComponent(f.name)}`}
                        download
                        style={{
                          background: "transparent",
                          border: "1px solid #d1d5db",
                          borderRadius: "6px",
                          padding: "0.3rem 0.6rem",
                          textDecoration: "none",
                          color: "#374151",
                          display: "inline-flex",
                        }}
                      >
                        <Download size={14} />
                      </a>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {previewName && (
        <div
          onClick={() => setPreviewName(null)}
          style={{
            position: "fixed",
            inset: 0,
            background: "rgba(0,0,0,0.6)",
            zIndex: 1000,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            padding: "2rem",
          }}
        >
          <div
            onClick={(e) => e.stopPropagation()}
            style={{
              background: "white",
              borderRadius: "12px",
              maxWidth: "900px",
              width: "100%",
              maxHeight: "85vh",
              display: "flex",
              flexDirection: "column",
            }}
          >
            <div
              style={{
                padding: "1rem 1.5rem",
                borderBottom: "1px solid #e5e7eb",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
              }}
            >
              <h3 style={{ fontWeight: 800 }}>{previewName}</h3>
              <button
                onClick={() => setPreviewName(null)}
                style={{
                  background: "transparent",
                  border: "none",
                  cursor: "pointer",
                }}
              >
                <X size={22} />
              </button>
            </div>
            <pre
              style={{
                flex: 1,
                overflow: "auto",
                padding: "1.5rem",
                background: "#1e1e1e",
                color: "#d4d4d4",
                margin: 0,
                fontSize: "0.8rem",
                borderRadius: "0 0 12px 12px",
              }}
            >
              {previewContent}
            </pre>
          </div>
        </div>
      )}
    </div>
  );
};

const LoginView = ({
  onLoginSuccess,
  onVisitorEntry,
  isDbActive,
}: {
  onLoginSuccess: () => void;
  onVisitorEntry: () => void;
  isDbActive: boolean;
}) => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await login({ username, password });
      onLoginSuccess();
    } catch {
      setError("Credenciales inválidas");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        height: "100vh",
        width: "100vw",
        background: "#111827",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <div
        className="dashboard-card"
        style={{ width: "420px", padding: "3rem" }}
      >
        <div style={{ textAlign: "center", marginBottom: "2.5rem" }}>
          <div
            className="sidebar-logo"
            style={{
              margin: "0 auto 1.5rem auto",
              width: "64px",
              height: "64px",
              fontSize: "1.8rem",
            }}
          >
            GS
          </div>
          <h1 style={{ fontSize: "1.8rem", fontWeight: 900 }}>goScadaSur</h1>
          <p style={{ color: "#6b7280", fontWeight: 600 }}>
            Gestión de Flujos y XML EPM
          </p>
        </div>

        {error && (
          <div
            style={{
              background: "#fef2f2",
              color: "#991b1b",
              padding: "0.75rem",
              borderRadius: "8px",
              marginBottom: "1.5rem",
              textAlign: "center",
              fontWeight: 700,
            }}
          >
            {error}
          </div>
        )}

        <form
          onSubmit={handleLogin}
          style={{ display: "flex", flexDirection: "column", gap: "1.25rem" }}
        >
          <div className="form-group">
            <label className="form-label" style={{ fontWeight: 700 }}>
              Usuario
            </label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              style={{
                width: "100%",
                padding: "0.85rem",
                borderRadius: "8px",
                border: "1px solid #d1d5db",
              }}
              placeholder="admin"
              disabled={!isDbActive}
            />
          </div>
          <div className="form-group">
            <label className="form-label" style={{ fontWeight: 700 }}>
              Contraseña
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              style={{
                width: "100%",
                padding: "0.85rem",
                borderRadius: "8px",
                border: "1px solid #d1d5db",
              }}
              placeholder="••••••••"
              disabled={!isDbActive}
            />
          </div>
          <button
            className="btn-epm"
            type="submit"
            disabled={loading || !isDbActive}
            style={{ height: "50px" }}
          >
            {loading ? (
              <Activity className="spinning" />
            ) : (
              "Ingresar al Sistema"
            )}
          </button>

          <div
            style={{
              margin: "1rem 0",
              display: "flex",
              alignItems: "center",
              gap: "1rem",
            }}
          >
            <div
              style={{ flex: 1, height: "1px", background: "#e5e7eb" }}
            ></div>
            <span
              style={{ fontSize: "0.8rem", color: "#9ca3af", fontWeight: 700 }}
            >
              O BIEN
            </span>
            <div
              style={{ flex: 1, height: "1px", background: "#e5e7eb" }}
            ></div>
          </div>

          <button
            className="nav-item"
            type="button"
            onClick={onVisitorEntry}
            style={{
              background: "#f3f4f6",
              color: "#374151",
              border: "1px solid #d1d5db",
              justifyContent: "center",
              height: "50px",
            }}
          >
            <ShieldCheck size={20} /> Ingresar como Visitante
          </button>
        </form>

        {!isDbActive && (
          <div
            style={{
              marginTop: "2rem",
              background: "#fffbeb",
              border: "1px solid #f59e0b",
              padding: "1rem",
              borderRadius: "8px",
              fontSize: "0.8rem",
              color: "#92400e",
              fontWeight: 700,
            }}
          >
            PostgreSQL no detectado. El inicio de sesión está deshabilitado. Use
            el modo visitante.
          </div>
        )}
      </div>
    </div>
  );
};

export default App;
