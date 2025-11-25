Run this prompt to re-generate:

> Create an ASCII-based visual document for all the data structures, data entities, and relationships between entities, within this small data analysis application. The ideal documentation will provide clarity on which app models/structs exist, their fields, and pointers to any other referenced struct. It is helpful to provide clarity between Go structs and internal models vs. SQLite relational DB models. 


╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║                           IoT DATA SANDBOX - ENTITY RELATIONSHIP DIAGRAM                         ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝

┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    LAYER ARCHITECTURE                                           │
├─────────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                                 │
│   ┌──────────────┐    ┌──────────────────┐    ┌────────────────┐    ┌───────────────────┐      │
│   │  HTTP/API    │───▶│    Services      │───▶│   Persistence  │───▶│  SQLite + Files   │      │
│   │  Handlers    │    │  (Business Logic)│    │   (Store)      │    │                   │      │
│   └──────────────┘    └──────────────────┘    └────────────────┘    └───────────────────┘      │
│         │                     │                       │                      │                  │
│         ▼                     ▼                       ▼                      ▼                  │
│   ┌──────────────┐    ┌──────────────────┐    ┌────────────────┐    ┌───────────────────┐      │
│   │ models.*     │    │ models.*         │    │ schemas.*      │    │ Tables:           │      │
│   │ (DTOs)       │    │ (Domain Entities)│    │ (DB Schemas)   │    │ projects          │      │
│   └──────────────┘    └──────────────────┘    └────────────────┘    │ data_sources      │      │
│                                                                      │ data_layers       │      │
│                                                                      └───────────────────┘      │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘


╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║                                GO APPLICATION MODELS (internal/models/)                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝

    ┌─────────────────────────────────────────┐
    │            models.Project               │
    ├─────────────────────────────────────────┤
    │  ProjectId    int64      [PK, persisted]│
    │  Name         string     [persisted]    │
    │  WhenCreated  time.Time  [persisted]    │
    ├ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┤
    │  Layers       []*DataLayer  [in-memory] │◀──────────────────────────────┐
    ├─────────────────────────────────────────┤                               │
    │  Methods:                               │                               │
    │  • GetLayerByName(name) *DataLayer      │                               │
    │  • GetLayerByIndex(idx) *DataLayer      │                               │
    │  • NumLayers() int                      │                               │
    └─────────────────────────────────────────┘                               │
                          │                                                   │
                          │ 1:N (Project has many Layers)                     │
                          ▼                                                   │
    ┌─────────────────────────────────────────┐                               │
    │            models.DataLayer             │                               │
    ├─────────────────────────────────────────┤                               │
    │  DataLayerId   int64     [PK, persisted]│                               │
    │  ProjectId     int64     [FK, persisted]│───────────────────────────────┘
    │  DataSourceId  *int64    [FK, persisted]│─────────────────────┐
    │  Name          string    [persisted]    │                     │
    │  Color         string    [persisted]    │                     │
    │  ZIndex        int       [persisted]    │                     │
    │  IsVisible     bool      [persisted]    │                     │
    ├ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┤                     │
    │  Project       *Project     [in-memory] │◀─────(back-ref)     │
    │  DataSource    *DataSource  [in-memory] │◀─────────┐          │
    ├─────────────────────────────────────────┤          │          │
    │  Methods:                               │          │          │
    │  • GetData() []DataEntry                │          │          │
    │  • GetTimeRange() (*time.Time, *time.Time)         │          │
    │  • IsHidden() bool                      │          │          │
    └─────────────────────────────────────────┘          │          │
                                                         │          │
                          ┌──────────────────────────────┘          │
                          │ N:1 (Many Layers can share 1 DataSource)│
                          ▼                                         │
    ┌─────────────────────────────────────────┐                     │
    │           models.DataSource             │                     │
    ├─────────────────────────────────────────┤                     │
    │  DataSourceId    int64    [PK, persisted]◀────────────────────┘
    │  Name            string   [persisted]   │
    │  DataSourceType  int      [persisted]   │ ─── 0="csv"
    │  DataSourcePath  string   [persisted]   │ ─── filename only (e.g., "data.csv")
    │  TimeLabel       string   [persisted]   │
    │  ValueLabel      string   [persisted]   │
    │  WhenCreated     time.Time [persisted]  │
    ├ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┤
    │  Data            []DataEntry [in-memory]│◀────────────────────┐
    ├─────────────────────────────────────────┤                     │
    │  Methods:                               │                     │
    │  • ToSchema() *DataSourceSchema         │                     │
    │  • FromSchema(schema)                   │                     │
    │  • GetRowCount() int                    │                     │
    │  • GetTimeRange() (*time.Time, *time.Time)                    │
    └─────────────────────────────────────────┘                     │
                                                                    │
                          ┌─────────────────────────────────────────┘
                          │ 1:N (DataSource has many DataEntries)
                          ▼
    ┌─────────────────────────────────────────┐
    │           models.DataEntry              │
    ├─────────────────────────────────────────┤
    │  Timestamp   time.Time    [in-memory]   │   ◄── NOT persisted in DB
    │  Value       float64      [in-memory]   │       Loaded from CSV files
    └─────────────────────────────────────────┘


    ┌─────────────────────────────────────────┐
    │          models.Integration             │
    ├─────────────────────────────────────────┤
    │  IntegrationId    int64   [PK]          │
    │  Name             string                │
    │  IntegrationType  int                   │ ─── 0="openai", 1="anthropic",
    │  HashedApiKey     *string               │     2="google", 3="ollama"
    └─────────────────────────────────────────┘


╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║                              DATABASE SCHEMA (internal/schemas/)                                 ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝

    ┌─────────────────────────────────────────┐
    │       schemas.DataSourceSchema          │
    ├─────────────────────────────────────────┤
    │  DataSourceId    int64      [PK]        │   ◄── Maps to data_sources table
    │  Name            string                 │
    │  DataSourceType  int                    │
    │  DataSourcePath  string                 │
    │  RowCount        int        [computed]  │   ◄── Derived from len(Data)
    │  StartTime       *time.Time [computed]  │   ◄── Derived from Data[0].Timestamp
    │  EndTime         *time.Time [computed]  │   ◄── Derived from Data[last].Timestamp
    │  TimeLabel       string                 │
    │  ValueLabel      string                 │
    │  WhenCreated     time.Time              │
    └─────────────────────────────────────────┘

                  ▲
                  │ Conversion
                  │
    ┌─────────────┴─────────────┐
    │  DataSource.ToSchema()    │  Go Model ──► Schema (for DB save)
    │  DataSource.FromSchema()  │  Schema   ──► Go Model (for DB load)
    └───────────────────────────┘


╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║                            SQLITE DATABASE TABLES (sandbox.db)                                   ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝

    ┌─────────────────────────────────────────────────────────────────────────────────────────────┐
    │                                    TABLE: projects                                          │
    ├────────────────┬──────────────────────────────┬──────────────────────────────────────────────┤
    │  Column        │  Type                        │  Constraints                                │
    ├────────────────┼──────────────────────────────┼──────────────────────────────────────────────┤
    │  project_id    │  INTEGER                     │  PRIMARY KEY AUTOINCREMENT                  │
    │  name          │  TEXT                        │  NOT NULL                                   │
    │  when_created  │  TIMESTAMP                   │  DEFAULT CURRENT_TIMESTAMP                  │
    └────────────────┴──────────────────────────────┴──────────────────────────────────────────────┘
                                            │
                                            │ 1:N
                                            ▼
    ┌─────────────────────────────────────────────────────────────────────────────────────────────┐
    │                                  TABLE: data_sources                                        │
    ├────────────────────┬──────────────────────────┬──────────────────────────────────────────────┤
    │  Column            │  Type                    │  Constraints                                │
    ├────────────────────┼──────────────────────────┼──────────────────────────────────────────────┤
    │  data_source_id    │  INTEGER                 │  PRIMARY KEY AUTOINCREMENT                  │
    │  project_id        │  INTEGER                 │  FK → projects(project_id) ON DELETE SET NULL│
    │  name              │  TEXT                    │  NOT NULL                                   │
    │  data_source_type  │  INTEGER                 │  NOT NULL                                   │
    │  data_source_path  │  TEXT                    │  NOT NULL                                   │
    │  row_count         │  INTEGER                 │  NOT NULL DEFAULT 0                         │
    │  start_time        │  TIMESTAMP               │  (nullable)                                 │
    │  end_time          │  TIMESTAMP               │  (nullable)                                 │
    │  time_label        │  TEXT                    │  NOT NULL DEFAULT 'time'                    │
    │  value_label       │  TEXT                    │  NOT NULL DEFAULT 'value'                   │
    │  when_created      │  TIMESTAMP               │  DEFAULT CURRENT_TIMESTAMP                  │
    └────────────────────┴──────────────────────────┴──────────────────────────────────────────────┘
                                            │
                                            │ 1:N
                                            ▼
    ┌─────────────────────────────────────────────────────────────────────────────────────────────┐
    │                                   TABLE: data_layers                                        │
    ├────────────────────┬──────────────────────────┬──────────────────────────────────────────────┤
    │  Column            │  Type                    │  Constraints                                │
    ├────────────────────┼──────────────────────────┼──────────────────────────────────────────────┤
    │  data_layer_id     │  INTEGER                 │  PRIMARY KEY AUTOINCREMENT                  │
    │  project_id        │  INTEGER                 │  FK → projects(project_id) ON DELETE CASCADE│
    │  data_source_id    │  INTEGER                 │  FK → data_sources(data_source_id) ON DELETE CASCADE│
    │  name              │  TEXT                    │  NOT NULL                                   │
    │  color             │  TEXT                    │  NOT NULL DEFAULT '#3b82f6'                 │
    │  z_index           │  INTEGER                 │  NOT NULL DEFAULT 0                         │
    │  is_visible        │  BOOLEAN                 │  NOT NULL DEFAULT 1                         │
    └────────────────────┴──────────────────────────┴──────────────────────────────────────────────┘

    INDEXES:
    ├── idx_data_sources_type      ON data_sources(data_source_type)
    ├── idx_data_sources_project   ON data_sources(project_id)
    ├── idx_data_layers_project    ON data_layers(project_id)
    └── idx_data_layers_z_index    ON data_layers(project_id, z_index)


╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║                                  ENTITY RELATIONSHIPS                                            ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝

    ┌───────────┐         1:N          ┌─────────────┐        N:1         ┌──────────────┐
    │  Project  │ ◄──────────────────▶ │  DataLayer  │ ─────────────────▶ │  DataSource  │
    └───────────┘                      └─────────────┘                    └──────────────┘
         │                                    │                                  │
         │                                    │                                  │
         │                                    │                                  ▼
         │                                    │                           ┌──────────────┐
         │                                    │                           │  DataEntry[] │
         │                                    │                           │  (in CSV file)│
         │                                    │                           └──────────────┘
         │                                    │
         ▼                                    ▼
    ┌─────────────────────────────────────────────────────────────────────────────────────┐
    │                              RELATIONSHIP RULES                                      │
    ├─────────────────────────────────────────────────────────────────────────────────────┤
    │  • A Project contains 0..N DataLayers (ordered by z_index)                          │
    │  • A DataLayer belongs to exactly 1 Project                                         │
    │  • A DataLayer optionally references 1 DataSource (can be NULL)                     │
    │  • Multiple DataLayers CAN share the same DataSource (N:1)                          │
    │  • A DataSource contains 0..N DataEntries (loaded from CSV, not in DB)              │
    │  • Deleting a Project CASCADE deletes all its DataLayers                            │
    │  • Deleting a Project sets data_sources.project_id to NULL (not deleted)            │
    │  • Deleting a DataSource CASCADE deletes all DataLayers referencing it              │
    └─────────────────────────────────────────────────────────────────────────────────────┘


╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║                              SUPPORTING INTERNAL STRUCTS                                         ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝

    ┌─────────────────────────────────────────┐
    │        timeseries.TimeSeriesData        │     (internal/timeseries/)
    ├─────────────────────────────────────────┤
    │  DataFrame   dataframe.DataFrame        │ ─── gota/dataframe (3rd party)
    │  StartTime   time.Time                  │
    │  EndTime     time.Time                  │
    │  RowCount    int                        │
    │  TimeLabel   string                     │
    │  ValueLabel  string                     │
    └─────────────────────────────────────────┘
         │
         │ Used during CSV loading/validation
         ▼
    ┌─────────────────────────────────────────┐
    │      timeseries.ValidationError         │
    ├─────────────────────────────────────────┤
    │  Message   string                       │
    └─────────────────────────────────────────┘


    ┌─────────────────────────────────────────┐
    │          storage.FileStore              │     (internal/storage/)
    ├─────────────────────────────────────────┤
    │  baseDir   string                       │ ─── "/opt/iot-data-sandbox" (darwin/linux)
    ├─────────────────────────────────────────┤     or ProgramData\iot-data-sandbox (windows)
    │  Methods:                               │     or ./app-files (fallback)
    │  • SaveFile(filename, reader, maxSize)  │
    │  • GetFilePath(filename) string         │
    │  • DeleteFile(filename) error           │
    │  • FileExists(filename) bool            │
    │  • GetBaseDir() string                  │
    └─────────────────────────────────────────┘


    ┌─────────────────────────────────────────┐
    │           tools.ToolManifest            │     (internal/tools/)
    ├─────────────────────────────────────────┤
    │  Name           string                  │
    │  Description    string                  │
    │  Category       ToolCategory            │ ─── "analysis"|"filter"|"transform"|"ai"|"other"
    │  Documentation  string                  │
    │  Parameters     []ParameterDefinition   │───┐
    │  Examples       []string                │   │
    └─────────────────────────────────────────┘   │
                                                  │
    ┌─────────────────────────────────────────┐   │
    │      tools.ParameterDefinition          │◀──┘
    ├─────────────────────────────────────────┤
    │  Name         string                    │
    │  Type         string                    │
    │  Description  string                    │
    │  Required     bool                      │
    └─────────────────────────────────────────┘


╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║                               REPOSITORY INTERFACES                                              ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝

    ┌──────────────────────────────────────────────────────────────────┐
    │         repositories.DataSourceRepository (interface)            │   (internal/repositories/)
    ├──────────────────────────────────────────────────────────────────┤
    │  Save(ds *schemas.DataSourceSchema) error                        │
    │  FindByID(id int64) (*schemas.DataSourceSchema, error)           │
    │  FindAll() ([]*schemas.DataSourceSchema, error)                  │
    │  Delete(id int64) error                                          │
    └──────────────────────────────────────────────────────────────────┘
                                    ▲
                                    │ implements
                                    │
    ┌──────────────────────────────────────────────────────────────────┐
    │               persistence.Store                                  │   (internal/persistence/)
    ├──────────────────────────────────────────────────────────────────┤
    │  db   *sql.DB                                                    │
    ├──────────────────────────────────────────────────────────────────┤
    │  // DataSource (via DataSourceRepository interface)              │
    │  Save(ds *schemas.DataSourceSchema)                              │
    │  FindByID(id) *schemas.DataSourceSchema                          │
    │  FindAll() []*schemas.DataSourceSchema                           │
    │  Delete(id)                                                      │
    ├──────────────────────────────────────────────────────────────────┤
    │  // Project operations                                           │
    │  SaveProject(p *models.Project)                                  │
    │  LoadProject(id) *models.Project                                 │
    │  LoadAllProjects() []*models.Project                             │
    │  DeleteProject(id)                                               │
    ├──────────────────────────────────────────────────────────────────┤
    │  // DataLayer operations                                         │
    │  SaveLayer(layer *models.DataLayer)                              │
    │  LoadLayer(id) *models.DataLayer                                 │
    │  LoadAllLayers() []*models.DataLayer                             │
    │  LoadLayersByProjectId(projectId) []*models.DataLayer            │
    │  LoadLayersByDataSourceId(dataSourceId) []*models.DataLayer      │
    │  LoadLayerWithDataSource(layerId) (*DataLayer, *Schema, error)   │
    │  LoadLayerWithProjectAndDataSource(layerId) (...)                │
    │  DeleteLayer(id)                                                 │
    └──────────────────────────────────────────────────────────────────┘


╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║                               SERVICE LAYER DEPENDENCIES                                         ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝

    ┌────────────────────┐
    │   ProjectService   │
    ├────────────────────┤        ┌──────────────────────┐
    │  store             │───────▶│  persistence.Store   │
    │  dataLayerService  │────┐   └──────────────────────┘
    └────────────────────┘    │              ▲
                              │              │
                              ▼              │
    ┌────────────────────┐    │              │
    │  DataLayerService  │◀───┘              │
    ├────────────────────┤                   │
    │  store             │───────────────────┤
    │  dataSourceService │────┐              │
    └────────────────────┘    │              │
                              │              │
                              ▼              │
    ┌────────────────────┐                   │
    │ DataSourceService  │                   │
    ├────────────────────┤                   │
    │  store             │───────────────────┘
    │  fileStore         │────────▶┌────────────────────┐
    └────────────────────┘         │  storage.FileStore │
                                   └────────────────────┘


╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║                              DATA FLOW: CSV UPLOAD TO VISUALIZATION                              ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝

    1. CSV Upload
       ───────────────────────────────────────────────────────────────────────────────────────────
       ┌─────────┐     ┌───────────────┐     ┌───────────────────┐     ┌──────────────────┐
       │  Client │────▶│  FileStore    │────▶│ timeseries.Load   │────▶│  TimeSeriesData  │
       │  (CSV)  │     │  .SaveFile()  │     │  AndValidateCSV() │     │  (validated)     │
       └─────────┘     └───────────────┘     └───────────────────┘     └──────────────────┘
                              │
                              ▼
                       ┌──────────────────┐
                       │  app-files/      │
                       │  data_123.csv    │
                       └──────────────────┘

    2. DataSource Creation
       ───────────────────────────────────────────────────────────────────────────────────────────
       ┌──────────────────┐     ┌───────────────────┐     ┌──────────────────┐
       │ TimeSeriesData   │────▶│ DataSourceService │────▶│ models.DataSource│
       │                  │     │ .CreateFromCSV()  │     │ (with Data[])    │
       └──────────────────┘     └───────────────────┘     └──────────────────┘
                                        │
                                        ▼
                                ┌───────────────────┐     ┌──────────────────┐
                                │ DataSource        │────▶│ DataSourceSchema │
                                │ .ToSchema()       │     │                  │
                                └───────────────────┘     └──────────────────┘
                                                                   │
                                                                   ▼
                                                          ┌──────────────────┐
                                                          │ SQLite DB        │
                                                          │ data_sources row │
                                                          └──────────────────┘

    3. Layer + DataSource Association
       ───────────────────────────────────────────────────────────────────────────────────────────
       ┌──────────────────┐     ┌───────────────────┐     ┌──────────────────┐
       │ models.DataLayer │────▶│ DataLayerService  │────▶│ SQLite DB        │
       │ .DataSourceId=X  │     │ .Save()           │     │ data_layers row  │
       └──────────────────┘     └───────────────────┘     └──────────────────┘

    4. Data Retrieval for Visualization
       ───────────────────────────────────────────────────────────────────────────────────────────
       ┌──────────────────┐     ┌───────────────────┐     ┌──────────────────┐
       │ API Request      │────▶│ DataLayerService  │────▶│ Store.LoadLayer  │
       │ GET /layer/{id}  │     │ .LoadWithDataSource()   │ WithDataSource() │
       └──────────────────┘     └───────────────────┘     └──────────────────┘
                                        │                          │
                                        │                          ▼
                                        │                 ┌──────────────────┐
                                        │                 │ DataSourceSchema │
                                        │                 └──────────────────┘
                                        │                          │
                                        ▼                          │
                                ┌───────────────────┐              │
                                │ loadDataFromCSV() │◀─────────────┘
                                │ (reads CSV file)  │
                                └───────────────────┘
                                        │
                                        ▼
                                ┌──────────────────┐
                                │ DataLayer with   │
                                │ .DataSource.Data │──────▶ Ready for visualization
                                │ []DataEntry      │
                                └──────────────────┘
