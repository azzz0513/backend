CREATE DATABASE IF NOT EXISTS checkin;
USE checkin;

CREATE TABLE IF NOT EXISTS `users` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `user_id` bigint(20) NOT NULL COMMENT '用户id',
    `username` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '用户名称',
    `password` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '用户密码',
    `email` varchar(64) COLLATE utf8mb4_general_ci COMMENT '用户邮箱',
    `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_username` (`username`) USING BTREE,
    UNIQUE KEY `idx_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `type` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `type_id` int(10) unsigned NOT NULL COMMENT '活动种类id',
    `type_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '种类名称',
    `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_type_id` (`type_id`),
    UNIQUE KEY `idx_type_name` (`type_name`)
)ENGINE=InnoDB DEFAULT  CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO `type` VALUES ('1', '1', '一次性签到', '2025-04-24 11:17:11', '2025-04-24 11:17:11');
INSERT INTO `type` VALUES ('2', '2', '长期考勤', '2025-04-24 11:17:11', '2025-04-24 11:17:11');

CREATE TABLE IF NOT EXISTS `way` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `way_id` int(10) unsigned NOT NULL COMMENT '打卡方式id',
    `way_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '打卡方式名称',
    `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_type_id` (`way_id`),
    UNIQUE KEY `idx_type_name` (`way_name`)
)ENGINE=InnoDB DEFAULT  CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO `way` VALUES ('1', '1', '验证码签到', '2025-04-24 11:17:11', '2025-04-24 11:17:11');

CREATE TABLE IF NOT EXISTS `member_list` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `list_id` bigint(20) NOT NULL COMMENT '列表id',
    `author_id` bigint(20) NOT NULL COMMENT '列表作者id',
    `member_count` int NOT NULL DEFAULT 0 COMMENT '列表成员数量',
    `list_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '列表名称',
    `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`list_id`),
    KEY `idx_list_id` (`list_id`),
    KEY `idx_author_id` (`author_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `list_participants` (
    `list_id` bigint(20) NOT NULL COMMENT '列表id（关联 member_list.list_id）',
    `user_id` bigint(20) NOT NULL COMMENT '参与用户id（关联 users.user_id）',
    PRIMARY KEY (`list_id`, `user_id`),
    FOREIGN KEY (`list_id`) REFERENCES `member_list`(`list_id`) ON DELETE CASCADE,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`user_id`) ON DELETE CASCADE,
    KEY `idx_member_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `checkins` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `checkin_id` bigint(20) NOT NULL COMMENT '活动id',
    `author_id` bigint(20) NOT NULL COMMENT '发起者的用户id',
    `title` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '标题',
    `content` varchar(8192) COLLATE utf8mb4_general_ci NOT NULL COMMENT '内容',
    `list_id` bigint(20) NOT NULL COMMENT '关联的用户列表',
    `type_id` bigint(20) NOT NULL COMMENT '所属活动种类',
    `way_id` bigint(20) NOT NULL COMMENT '打卡方式',
    `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '活动状态',
    `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '活动开始时间',
    `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '活动更新时间',

    -- 验证码打卡专用字段
    `password` varchar(64) COLLATE utf8mb4_general_ci COMMENT '验证码',

    -- 长期考勤专用字段
    `start_date` DATE NULL COMMENT '开始日期（长期活动必填）',
    `end_date` DATE NULL COMMENT '结束日期（长期活动必填）',
    `daily_deadline` TIME NULL COMMENT '每日截止时间（格式：HH:mm:ss）',

    -- 一次性签到专用字段
    `start_time` DATETIME NULL COMMENT '开始时间（单次活动必填）',
    `duration_minutes` INT UNSIGNED NULL COMMENT '持续时间（分钟）',

    PRIMARY KEY (`id`),
    FOREIGN KEY (`list_id`) REFERENCES `member_list`(`list_id`) ON DELETE CASCADE,
    UNIQUE KEY `idx_checkin_id` (`checkin_id`),
    KEY `idx_author_id` (`author_id`),
    KEY `idx_type_id` (`type_id`),
    KEY `idx_list_id` (`list_id`),
    KEY `idx_way_id` (`way_id`),

    -- 活动种类约束条件
    CONSTRAINT chk_single_event CHECK (
    (type_id != 1) OR (  -- type_id=1是一次性签到类型
                          start_time IS NOT NULL
                          AND duration_minutes IS NOT NULL
                          AND start_date IS NULL
                          AND end_date IS NULL
                          AND daily_deadline IS NULL
                      )
    ),
    CONSTRAINT chk_long_term CHECK (
        (type_id != 2) OR (  -- type_id=2是长期考勤类型
            start_date IS NOT NULL
            AND end_date IS NOT NULL
            AND daily_deadline IS NOT NULL
            AND start_time IS NULL
            AND duration_minutes IS NULL
            )
    ),

    -- 活动打卡方式约束条件
    CONSTRAINT chk_password CHECK (
        (way_id != 1) OR (  -- way_id=1是验证码签到类型
            password IS NOT NULL
        )
    ),
    CONSTRAINT chk_date_order CHECK (
        start_date <= end_date
    ),
    CONSTRAINT chk_positive_duration CHECK (
        duration_minutes > 0
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `checkin_records` (
    checkin_id BIGINT(20) NOT NULL COMMENT '关联checkins.checkin_id',
    list_id BIGINT(20) NOT NULL COMMENT '关联member_list.list',
    user_id BIGINT(20) NOT NULL COMMENT '关联list_participants.user_id',
    is_checked TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已打卡（0-未打，1-已打）',
    check_time timestamp NULL COMMENT '最后一次打卡时间',
    PRIMARY KEY (checkin_id, user_id),
    FOREIGN KEY (checkin_id) REFERENCES checkins(checkin_id) ON DELETE CASCADE,
    FOREIGN KEY (list_id) REFERENCES member_list(list_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES list_participants(user_id) ON DELETE CASCADE,
    INDEX idx_check_time (check_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `checkin_stats` (
    stat_id BIGINT(20) AUTO_INCREMENT PRIMARY KEY,
    checkin_id BIGINT(20) NOT NULL COMMENT '关联的打卡活动ID',
    user_id BIGINT(20) NOT NULL COMMENT '用户ID',
    period_type ENUM('day','week','month','year') NOT NULL COMMENT '统计周期类型',
    period_start DATE NOT NULL COMMENT '统计周期开始日期',
    period_end DATE NOT NULL COMMENT '统计周期结束日期',
    checkin_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打卡次数',
    last_checkin_time DATETIME COMMENT '最后一次打卡时间',

    UNIQUE KEY idx_unique_stat (checkin_id, user_id, period_type, period_start),
    KEY idx_period_range (period_start, period_end),
    FOREIGN KEY (checkin_id) REFERENCES checkins(checkin_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
