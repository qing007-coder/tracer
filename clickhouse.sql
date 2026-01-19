-- 1. 创建数据库
CREATE DATABASE IF NOT EXISTS tracer;

-- 2. 创建表
CREATE TABLE IF NOT EXISTS tracer.trace_spans (
    -- 基础关联字段
    TraceID          String,
    SpanID           String,
    ParentSpanID     String,

    -- 核心维度 (使用 LowCardinality 优化低基数字符串，大幅提升查询性能)
                                                  ServiceName      LowCardinality(String),
    OperationName    LowCardinality(String),
    Kind             String,

    -- 时间与性能
    StartTimeUs      Int64,
    DurationUs       Int64,
    TimestampUs      Int64,

    -- 状态信息
    StatusCode       LowCardinality(String),
    StatusMessage    String,

    -- 扩展信息 (Map 类型，非常适合 Tracing)
    Attributes       Map(String, String),
    ResourceAttrs    Map(String, String),

    -- 事件信息 (使用数组存储)
    EventNames       Array(String),
    EventTimesUs     Array(Int64),
    EventAttrs       Array(String), -- 存储 JSON 字符串数组

-- 其他信息
    ProcessID        String

    ) ENGINE = MergeTree()
-- 关键配置 1: 按天分区，方便后期按天物理删除旧数据
    PARTITION BY toYYYYMMDD(toDateTime64(TimestampUs/1000000, 6))
-- 关键配置 2: 排序键，决定了数据在磁盘上的存放顺序，也是查询索引
    ORDER BY (ServiceName, OperationName, TimestampUs)
-- 关键配置 3: 设置数据保存时间 (例如保存 7 天)
    TTL toDateTime(TimestampUs / 1000000) + INTERVAL 7 DAY
