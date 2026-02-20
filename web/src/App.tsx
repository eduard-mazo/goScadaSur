import { useState, useEffect, useMemo } from 'react';
import { 
  LayoutDashboard, Search, Database, FileCode, Settings, Activity, AlertTriangle,
  ChevronRight, Save, RefreshCw, Plus, Trash2,
  CheckCircle2, XCircle, Terminal, Code, List, Edit3, Copy, Search as SearchIcon
} from 'lucide-react';
import { 
  getAppConfig, saveAppConfig, getTemplateStats, getDasipConfig, saveDasipConfig, 
  getRawTemplates, saveRawTemplates,
  type AppConfig, type TemplateStats, type DasipConfig 
} from './api';

type View = 'dashboard' | 'search' | 'query' | 'generator' | 'settings';

function App() {
  const [currentView, setCurrentView] = useState<View>('dashboard');
  const [appConfig, setAppConfig] = useState<AppConfig | null>(null);
  const [templateStats, setTemplateStats] = useState<TemplateStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [isConnected, setIsConnected] = useState(false);

  const fetchData = async () => {
    try {
      const configData = await getAppConfig();
      const stats = await getTemplateStats();
      setAppConfig(configData);
      setTemplateStats(stats);
      setIsConnected(true);
    } catch (error) {
      console.error('Error fetching initial data:', error);
      setIsConnected(false);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 30000);
    return () => clearInterval(interval);
  }, []);

  const renderView = () => {
    switch (currentView) {
      case 'dashboard':
        return <DashboardView stats={templateStats} onRefresh={fetchData} />;
      case 'search':
        return <SearchView />;
      case 'query':
        return <QueryView />;
      case 'generator':
        return <GeneratorView />;
      case 'settings':
        return <SettingsView config={appConfig} onUpdate={fetchData} />;
      default:
        return <DashboardView stats={templateStats} onRefresh={fetchData} />;
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
          <button className={`nav-item ${currentView === 'dashboard' ? 'active' : ''}`} onClick={() => setCurrentView('dashboard')}>
            <LayoutDashboard size={20} /><span>Dashboard</span>
          </button>
          <button className={`nav-item ${currentView === 'search' ? 'active' : ''}`} onClick={() => setCurrentView('search')}>
            <Search size={20} /><span>Station Search</span>
          </button>
          <button className={`nav-item ${currentView === 'query' ? 'active' : ''}`} onClick={() => setCurrentView('query')}>
            <Database size={20} /><span>Direct Query</span>
          </button>
          <button className={`nav-item ${currentView === 'generator' ? 'active' : ''}`} onClick={() => setCurrentView('generator')}>
            <FileCode size={20} /><span>XML Generator</span>
          </button>
        </nav>
        <nav className="sidebar-nav" style={{ borderTop: '1px solid rgba(255,255,255,0.1)' }}>
          <button className={`nav-item ${currentView === 'settings' ? 'active' : ''}`} onClick={() => setCurrentView('settings')}>
            <Settings size={20} /><span>Settings</span>
          </button>
        </nav>
      </aside>

      <main className="main-content">
        <header className="header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <span style={{ color: 'var(--text-secondary)' }}>System</span>
            <ChevronRight size={16} color="var(--text-muted)" />
            <span style={{ fontWeight: 700, color: 'var(--epm-forest-green)' }}>{currentView.charAt(0).toUpperCase() + currentView.slice(1)}</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <div className={`badge ${isConnected ? 'success' : 'error'}`} style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
              {isConnected ? <CheckCircle2 size={14} /> : <XCircle size={14} />}
              {isConnected ? 'Connected' : 'Disconnected'}
            </div>
            <Activity size={20} color={isConnected ? "#0d9648" : "#ef4444"} className={loading ? "spinning" : ""} />
          </div>
        </header>

        <div className="view-container">
          {loading && !appConfig ? (
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '60vh', gap: '1rem' }}>
              <Activity size={48} className="spinning" color="#0d9648" />
              <p style={{ fontWeight: 600, color: 'var(--text-secondary)' }}>Initializing System...</p>
            </div>
          ) : renderView()}
        </div>
      </main>
    </div>
  );
}

const DashboardView = ({ stats, onRefresh }: { stats: TemplateStats | null, onRefresh: () => void }) => (
  <div>
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
      <div>
        <h2 className="card-title" style={{ marginBottom: '0.25rem' }}>System Overview</h2>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>Real-time statistics from current configuration</p>
      </div>
      <button className="btn-outline" onClick={onRefresh} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        <RefreshCw size={16} /> Update Data
      </button>
    </div>
    <div className="stats-grid">
      <div className="stat-card">
        <div className="stat-value">{stats?.stats.total || 0}</div>
        <div className="stat-label">Total Templates</div>
      </div>
      <div className="stat-card"><div className="stat-value">{stats?.stats.analog || 0}</div><div className="stat-label">Analog Definitions</div></div>
      <div className="stat-card"><div className="stat-value">{stats?.stats.discrete || 0}</div><div className="stat-label">Discrete Definitions</div></div>
      <div className="stat-card"><div className="stat-value">{stats?.stats.breaker || 0}</div><div className="stat-label">Breaker Definitions</div></div>
    </div>
    <div style={{ marginTop: '2.5rem' }}>
      <div className="card">
        <div className="card-title"><AlertTriangle size={20} color="#f59e0b" /> Configuration Health</div>
        <div style={{ color: 'var(--text-secondary)', fontSize: '0.95rem' }}>
          {stats?.warnings && stats.warnings.length > 0 ? (
            <ul style={{ paddingLeft: '1.5rem', marginTop: '1rem' }}>
              {stats.warnings.map((w, i) => <li key={i} style={{ marginBottom: '0.75rem', color: 'var(--text-primary)' }}>{w}</li>)}
            </ul>
          ) : <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--success)', fontWeight: 600, marginTop: '1rem' }}>
                <CheckCircle2 size={18} /> All templates passed validation.
              </div>}
        </div>
      </div>
    </div>
  </div>
);

const SettingsView = ({ config, onUpdate }: { config: AppConfig | null, onUpdate: () => void }) => {
  const [activeTab, setActiveTab] = useState<'general' | 'dasip' | 'templates'>('general');
  const [localConfig, setLocalConfig] = useState<AppConfig | null>(null);
  const [dasip, setDasip] = useState<DasipConfig | null>(null);
  const [rawTemplates, setRawTemplates] = useState<string>('');
  const [templateEditorMode, setTemplateEditorMode] = useState<'form' | 'json'>('form');
  const [selectedTemplateKey, setSelectedTemplateKey] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [saving, setSaving] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (config) setLocalConfig(JSON.parse(JSON.stringify(config)));
  }, [config]);

  useEffect(() => {
    const loadTabData = async () => {
      if (activeTab === 'general') return;
      setLoading(true);
      try {
        if (activeTab === 'dasip') {
          const data = await getDasipConfig();
          setDasip(data);
        } else if (activeTab === 'templates') {
          const data = await getRawTemplates();
          setRawTemplates(data);
        }
      } catch (e) {
        console.error('Error loading tab data', e);
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
      .filter(key => key.toLowerCase().includes(searchTerm.toLowerCase()))
      .sort();
  }, [parsedTemplates, searchTerm]);

  const handleSaveGeneral = async () => {
    if (!localConfig) return;
    setSaving(true);
    try {
      await saveAppConfig(localConfig);
      alert('Application configuration saved successfully');
      onUpdate();
    } catch (e) {
      alert('Error saving general configuration');
    } finally {
      setSaving(false);
    }
  };

  const handleSaveDasip = async () => {
    if (!dasip) return;
    setSaving(true);
    try {
      await saveDasipConfig(dasip);
      alert('DASIP mapping saved successfully');
      onUpdate();
    } catch (e) {
      alert('Error saving DASIP');
    } finally {
      setSaving(false);
    }
  };

  const handleSaveTemplates = async (content: string) => {
    setSaving(true);
    try {
      await saveRawTemplates(content);
      setRawTemplates(content);
      alert('Templates saved successfully');
      onUpdate();
    } catch (e) {
      alert('Error saving templates. Check JSON format.');
    } finally {
      setSaving(false);
    }
  };

  const handleTemplateChange = (key: string, type: string, field: string, value: any, subfield?: string) => {
    const updated = { ...parsedTemplates };
    if (subfield) {
      if (!updated[key][type][field]) updated[key][type][field] = {};
      updated[key][type][field][subfield] = value;
    } else {
      updated[key][type][field] = value;
    }
    setRawTemplates(JSON.stringify(updated, null, 2));
  };

  return (
    <div className="card">
      <div style={{ marginBottom: '2rem' }}>
        <h2 className="card-title"><Settings size={20} /> System Configuration</h2>
        <div className="settings-tabs">
          <button className={`tab ${activeTab === 'general' ? 'active' : ''}`} onClick={() => setActiveTab('general')}>App Engine</button>
          <button className={`tab ${activeTab === 'dasip' ? 'active' : ''}`} onClick={() => setActiveTab('dasip')}>Network Map</button>
          <button className={`tab ${activeTab === 'templates' ? 'active' : ''}`} onClick={() => setActiveTab('templates')}>XML Templates</button>
        </div>
      </div>

      {loading ? (
        <div style={{ padding: '4rem', textAlign: 'center' }}><Activity className="spinning" color="#0d9648" /></div>
      ) : (
        <>
          {activeTab === 'general' && localConfig && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
              <div className="stats-grid" style={{ gridTemplateColumns: '1fr 1fr' }}>
                <div className="card" style={{ background: 'var(--bg-tertiary)', border: 'none' }}>
                  <h3 style={{ fontSize: '0.9rem', fontWeight: 700, marginBottom: '1.5rem', color: 'var(--epm-forest-green)' }}>Application Context</h3>
                  <div className="form-group">
                    <label className="form-label">System Name</label>
                    <input type="text" value={localConfig.app.name} onChange={(e) => setLocalConfig({...localConfig, app: {...localConfig.app, name: e.target.value}})} style={{ width: '100%' }} />
                  </div>
                  <div className="form-group">
                    <label className="form-label">Description</label>
                    <textarea value={localConfig.app.description} onChange={(e) => setLocalConfig({...localConfig, app: {...localConfig.app, description: e.target.value}})} style={{ width: '100%', height: '80px' }} />
                  </div>
                </div>
                <div className="card" style={{ background: 'var(--bg-tertiary)', border: 'none' }}>
                  <h3 style={{ fontSize: '0.9rem', fontWeight: 700, marginBottom: '1.5rem', color: 'var(--epm-forest-green)' }}>Database Engine</h3>
                  <div className="form-group">
                    <label className="form-label">C# Connector Path</label>
                    <input type="text" value={localConfig.database.csharp_executable} onChange={(e) => setLocalConfig({...localConfig, database: {...localConfig.database, csharp_executable: e.target.value}})} style={{ width: '100%' }} />
                  </div>
                  <div className="form-group">
                    <label className="form-label">Connection Timeout (ms)</label>
                    <input type="number" value={localConfig.database.connection_timeout} onChange={(e) => setLocalConfig({...localConfig, database: {...localConfig.database, connection_timeout: parseInt(e.target.value)}})} style={{ width: '100%' }} />
                  </div>
                </div>
                <div className="card" style={{ background: 'var(--bg-tertiary)', border: 'none' }}>
                  <h3 style={{ fontSize: '0.9rem', fontWeight: 700, marginBottom: '1.5rem', color: 'var(--epm-forest-green)' }}>Internal File Paths</h3>
                  <div className="form-group">
                    <label className="form-label">Templates Library (.json)</label>
                    <input type="text" value={localConfig.files.templates} onChange={(e) => setLocalConfig({...localConfig, files: {...localConfig.files, templates: e.target.value}})} style={{ width: '100%' }} />
                  </div>
                  <div className="form-group">
                    <label className="form-label">Network Mapping (.yaml)</label>
                    <input type="text" value={localConfig.files.dasip_mapping} onChange={(e) => setLocalConfig({...localConfig, files: {...localConfig.files, dasip_mapping: e.target.value}})} style={{ width: '100%' }} />
                  </div>
                </div>
                <div className="card" style={{ background: 'var(--bg-tertiary)', border: 'none' }}>
                  <h3 style={{ fontSize: '0.9rem', fontWeight: 700, marginBottom: '1.5rem', color: 'var(--epm-forest-green)' }}>Processing Engine</h3>
                  <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', marginBottom: '1.5rem' }}>
                    <input type="checkbox" checked={localConfig.processing.parallel_enabled} onChange={(e) => setLocalConfig({...localConfig, processing: {...localConfig.processing, parallel_enabled: e.target.checked}})} />
                    <label className="form-label" style={{ marginBottom: 0 }}>Enable Parallel Multi-threading</label>
                  </div>
                  <div className="form-group">
                    <label className="form-label">Max Worker Threads</label>
                    <input type="number" value={localConfig.processing.max_workers} onChange={(e) => setLocalConfig({...localConfig, processing: {...localConfig.processing, max_workers: parseInt(e.target.value)}})} style={{ width: '100%' }} />
                  </div>
                </div>
              </div>
              <button className="btn-primary" onClick={handleSaveGeneral} disabled={saving} style={{ width: 'fit-content' }}>
                <Save size={16} /> {saving ? 'Saving...' : 'Commit App Configuration'}
              </button>
            </div>
          )}

          {activeTab === 'dasip' && (
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                <h3 style={{ fontSize: '1rem', fontWeight: 700 }}>Network Infrastructure Map</h3>
                <button className="btn-primary" onClick={handleSaveDasip} disabled={saving} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <Save size={16} /> {saving ? 'Saving...' : 'Apply Network Map'}
                </button>
              </div>
              <div className="form-group" style={{ background: 'var(--bg-tertiary)', padding: '1.5rem', borderRadius: 'var(--radius-md)' }}>
                <label className="form-label">Global Default Network Path</label>
                <input type="text" value={dasip?.default_path || ''} onChange={(e) => setDasip(d => d ? { ...d, default_path: e.target.value } : null)} style={{ width: '100%', fontSize: '1rem', fontWeight: 600 }} />
              </div>
              <div className="data-table-container">
                <table className="data-table">
                  <thead><tr><th>DASIP Identifier (IP/Key)</th><th>Survalent Network Path</th><th style={{ width: '60px' }}>Action</th></tr></thead>
                  <tbody>
                    {Object.entries(dasip?.dasip_mapping || {}).map(([key, val]) => (
                      <tr key={key}>
                        <td style={{ fontWeight: 700, color: 'var(--epm-forest-green)' }}>{key}</td>
                        <td><input type="text" value={val} onChange={(e) => {
                          const newMapping = { ...dasip!.dasip_mapping, [key]: e.target.value };
                          setDasip({ ...dasip!, dasip_mapping: newMapping });
                        }} style={{ width: '100%', background: 'transparent', border: 'none' }} /></td>
                        <td style={{ textAlign: 'center' }}><button onClick={() => {
                          const newMapping = { ...dasip!.dasip_mapping };
                          delete newMapping[key];
                          setDasip({ ...dasip!, dasip_mapping: newMapping });
                        }} className="btn-icon"><Trash2 size={18} color="#ef4444" /></button></td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <button className="btn-outline" onClick={() => {
                const key = prompt('Enter new DASIP Key (e.g. 192.168.1.50)');
                if (key) setDasip(d => d ? { ...d, dasip_mapping: { ...d.dasip_mapping, [key]: '' } } : null);
              }} style={{ marginTop: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Plus size={16} /> Add Entry
              </button>
            </div>
          )}

          {activeTab === 'templates' && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <h3 style={{ fontSize: '1rem', fontWeight: 700 }}>XML Signal Templates</h3>
                  <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>Map physical elements to SCADA XML structures</p>
                </div>
                <div style={{ display: 'flex', gap: '0.75rem' }}>
                  <div className="settings-tabs" style={{ marginBottom: 0, padding: '0.2rem' }}>
                    <button className={`tab ${templateEditorMode === 'form' ? 'active' : ''}`} onClick={() => setTemplateEditorMode('form')} style={{ padding: '0.4rem 0.8rem' }}><List size={14} /></button>
                    <button className={`tab ${templateEditorMode === 'json' ? 'active' : ''}`} onClick={() => setTemplateEditorMode('json')} style={{ padding: '0.4rem 0.8rem' }}><Code size={14} /></button>
                  </div>
                  <button className="btn-primary" onClick={() => handleSaveTemplates(rawTemplates)} disabled={saving} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <Save size={16} /> {saving ? 'Deploying...' : 'Deploy Changes'}
                  </button>
                </div>
              </div>

              {templateEditorMode === 'json' ? (
                <textarea 
                  value={rawTemplates} 
                  onChange={(e) => setRawTemplates(e.target.value)}
                  spellCheck={false}
                  style={{ 
                    width: '100%', height: '600px', fontFamily: 'var(--font-mono)', 
                    fontSize: '0.95rem', background: '#1e1e1e', color: '#d4d4d4',
                    padding: '1.5rem', lineHeight: '1.5', borderRadius: 'var(--radius-md)',
                    boxShadow: 'inset 0 2px 10px rgba(0,0,0,0.2)'
                  }}
                />
              ) : (
                <div style={{ display: 'grid', gridTemplateColumns: '320px 1fr', gap: '2rem', minHeight: '600px' }}>
                  <div className="card" style={{ background: 'var(--bg-tertiary)', border: 'none', padding: '1rem', overflowY: 'auto', maxHeight: '700px' }}>
                    <div className="search-container">
                      <SearchIcon size={16} className="search-icon" />
                      <input 
                        type="text" 
                        placeholder="Search templates..." 
                        className="search-input"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                      />
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', padding: '0 0.5rem' }}>
                      <span style={{ fontWeight: 700, fontSize: '0.75rem', color: 'var(--text-muted)' }}>ELEMENTS ({filteredTemplateKeys.length})</span>
                      <button className="btn-icon" onClick={() => {
                        const name = prompt('New Element Key (e.g. PT_KV)');
                        if (name) {
                          const updated = { ...parsedTemplates, [name]: { Analog: { Name: name } } };
                          setRawTemplates(JSON.stringify(updated, null, 2));
                          setSelectedTemplateKey(name);
                        }
                      }} title="Add New Template"><Plus size={16} color="var(--epm-forest-green)" /></button>
                    </div>
                    {filteredTemplateKeys.map(key => (
                      <div key={key} className={`template-list-item ${selectedTemplateKey === key ? 'selected' : ''}`} onClick={() => setSelectedTemplateKey(key)}>
                        <span style={{ fontWeight: 600, fontSize: '0.9rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{key}</span>
                        <div className="actions">
                          <button className="btn-icon" onClick={(e) => {
                            e.stopPropagation();
                            const newKey = key + "_COPY";
                            const updated = { ...parsedTemplates, [newKey]: JSON.parse(JSON.stringify(parsedTemplates[key])) };
                            setRawTemplates(JSON.stringify(updated, null, 2));
                            setSelectedTemplateKey(newKey);
                          }} title="Clone"><Copy size={14} /></button>
                          <button className="btn-icon" onClick={(e) => {
                            e.stopPropagation();
                            if (confirm('Delete this template?')) {
                              const updated = { ...parsedTemplates };
                              delete updated[key];
                              setRawTemplates(JSON.stringify(updated, null, 2));
                              if (selectedTemplateKey === key) setSelectedTemplateKey(null);
                            }
                          }} title="Delete"><Trash2 size={14} /></button>
                        </div>
                      </div>
                    ))}
                  </div>
                  
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                    {selectedTemplateKey && parsedTemplates[selectedTemplateKey] ? (
                      <div className="card" style={{ borderStyle: 'dashed', flex: 1 }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
                          <h3 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--epm-forest-green)', fontFamily: 'var(--font-title)' }}>
                            {selectedTemplateKey}
                          </h3>
                          <div className="badge success">Active Element</div>
                        </div>
                        
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
                          {['Analog', 'Discrete', 'Breaker'].map(type => {
                            const isTypeActive = !!parsedTemplates[selectedTemplateKey][type];
                            return (
                              <div key={type} className="property-section" style={{ opacity: isTypeActive ? 1 : 0.6 }}>
                                <div className="property-section-header">
                                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                                    <input 
                                      type="checkbox" 
                                      checked={isTypeActive} 
                                      onChange={(e) => {
                                        const updated = { ...parsedTemplates };
                                        if (e.target.checked) updated[selectedTemplateKey][type] = { Name: selectedTemplateKey };
                                        else delete updated[selectedTemplateKey][type];
                                        setRawTemplates(JSON.stringify(updated, null, 2));
                                      }}
                                    />
                                    <span className="property-section-title">{type} Configuration</span>
                                  </div>
                                  {isTypeActive && <Edit3 size={14} color="var(--epm-forest-green)" />}
                                </div>

                                {isTypeActive && (
                                  <div className="property-grid">
                                    <div className="form-group">
                                      <label className="form-label">Display Name</label>
                                      <input type="text" value={parsedTemplates[selectedTemplateKey][type].Name || ''} onChange={(e) => handleTemplateChange(selectedTemplateKey, type, 'Name', e.target.value)} style={{ width: '100%' }} />
                                    </div>
                                    <div className="form-group">
                                      <label className="form-label">Element Type</label>
                                      <input type="text" value={parsedTemplates[selectedTemplateKey][type].ElementType || ''} onChange={(e) => handleTemplateChange(selectedTemplateKey, type, 'ElementType', e.target.value)} style={{ width: '100%' }} />
                                    </div>
                                    {type === 'Analog' && (
                                      <>
                                        <div className="form-group">
                                          <label className="form-label">Unit of Measure</label>
                                          <input type="text" value={parsedTemplates[selectedTemplateKey][type].UnitOfMeasure || ''} onChange={(e) => handleTemplateChange(selectedTemplateKey, type, 'UnitOfMeasure', e.target.value)} style={{ width: '100%' }} />
                                        </div>
                                        <div className="nested-property">
                                          <span style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-muted)' }}>ANALOG VALUE</span>
                                          <div style={{ marginTop: '0.5rem', display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                                            <div className="form-group">
                                              <label className="form-label">Archive ID</label>
                                              <input type="text" value={parsedTemplates[selectedTemplateKey][type].AnalogValue?.Archive || ''} onChange={(e) => handleTemplateChange(selectedTemplateKey, type, 'AnalogValue', e.target.value, 'Archive')} style={{ width: '100%' }} />
                                            </div>
                                            <div className="form-group">
                                              <label className="form-label">Info Name</label>
                                              <input type="text" value={parsedTemplates[selectedTemplateKey][type].AnalogValue?.InfoName || ''} onChange={(e) => handleTemplateChange(selectedTemplateKey, type, 'AnalogValue', e.target.value, 'InfoName')} style={{ width: '100%' }} />
                                            </div>
                                          </div>
                                        </div>
                                      </>
                                    )}
                                    {type === 'Discrete' && (
                                      <div className="nested-property">
                                        <span style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-muted)' }}>DISCRETE VALUE</span>
                                        <div style={{ marginTop: '0.5rem' }}>
                                          <div className="form-group">
                                            <label className="form-label">Info Name</label>
                                            <input type="text" value={parsedTemplates[selectedTemplateKey][type].DiscreteValue?.InfoName || ''} onChange={(e) => handleTemplateChange(selectedTemplateKey, type, 'DiscreteValue', e.target.value, 'InfoName')} style={{ width: '100%' }} />
                                          </div>
                                        </div>
                                      </div>
                                    )}
                                  </div>
                                )}
                              </div>
                            );
                          })}
                        </div>
                        
                        <div style={{ marginTop: '2rem' }}>
                          <span style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-muted)', marginBottom: '0.5rem', display: 'block' }}>LOCAL JSON SNIPPET</span>
                          <div className="code-preview-panel">
                            {JSON.stringify(parsedTemplates[selectedTemplateKey], null, 2)}
                          </div>
                        </div>
                      </div>
                    ) : (
                      <div className="empty-state">
                        <Edit3 size={64} style={{ marginBottom: '1.5rem', opacity: 0.1 }} />
                        <h3 style={{ fontSize: '1.25rem', fontWeight: 700 }}>No element selected</h3>
                        <p>Pick a template from the element list to edit its properties or add a new one.</p>
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

const SearchView = () => (
  <div>
    <div className="card">
      <h2 className="card-title"><Search size={20} /> Search Station Signals</h2>
      <div className="stats-grid" style={{ gridTemplateColumns: '1fr 1fr' }}>
        <div className="form-group"><label className="form-label">Database Host</label><input type="text" placeholder="e.g. 10.0.0.1" style={{ width: '100%' }} /></div>
        <div className="form-group"><label className="form-label">System Path</label><input type="text" placeholder="e.g. EMPRESA/REGION/B1/B2/B3" style={{ width: '100%' }} /></div>
      </div>
      <div className="stats-grid" style={{ gridTemplateColumns: '1fr 1fr 1fr' }}>
        <div className="form-group"><label className="form-label">Username</label><input type="text" placeholder="DB User" style={{ width: '100%' }} /></div>
        <div className="form-group"><label className="form-label">Password</label><input type="password" placeholder="••••••••" style={{ width: '100%' }} /></div>
        <div className="form-group"><label className="form-label">Area of Responsibility</label><input type="text" placeholder="AOR" style={{ width: '100%' }} /></div>
      </div>
      <button className="btn-primary" style={{ marginTop: '1rem' }}>Run Search Operation</button>
    </div>
  </div>
);

const QueryView = () => (
  <div className="card" style={{ height: 'calc(100vh - 12rem)', display: 'flex', flexDirection: 'column' }}>
    <h2 className="card-title"><Terminal size={20} /> Advanced SQL Engine</h2>
    <div className="form-group" style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
      <label className="form-label">SQL Script</label>
      <textarea 
        style={{ flex: 1, width: '100%', fontFamily: 'var(--font-mono)', fontSize: '1rem', resize: 'none', padding: '1.5rem', background: 'var(--bg-tertiary)' }} 
        placeholder="SELECT * FROM Signals WHERE Station = '...' " 
      />
    </div>
    <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem' }}>
      <button className="btn-primary">Execute Query</button>
      <button className="btn-outline">Clear Buffer</button>
    </div>
  </div>
);

const GeneratorView = () => (
  <div>
    <div className="card">
      <h2 className="card-title"><FileCode size={20} /> IFS/IMM Processor</h2>
      <div style={{ border: '3px dashed var(--border)', borderRadius: 'var(--radius-lg)', padding: '5rem 2rem', textAlign: 'center', background: 'var(--bg-tertiary)', cursor: 'pointer' }}>
        <FileCode size={64} color="#0d9648" style={{ marginBottom: '1.5rem', opacity: 0.5 }} />
        <p style={{ fontWeight: 700, fontSize: '1.25rem', color: 'var(--text-primary)' }}>Upload Data File</p>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.95rem', marginTop: '0.5rem' }}>Supported formats: .csv, .xlsx, .xls</p>
        <button className="btn-primary" style={{ marginTop: '2rem' }}>Select Local File</button>
      </div>
    </div>
  </div>
);

export default App;
