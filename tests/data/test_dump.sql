-- ============================================
-- LVAN Dumper SQL 导入功能测试数据
-- ============================================
-- 此文件包含用于测试 import-sql 功能的测试数据
-- 覆盖所有 MySQL 数据类型和边界条件
-- ============================================

-- MySQL dump 10.13  Distrib 8.0.28, for Linux (x86_64)
--
-- Host: localhost    Database: lvan_dumper_test
-- ------------------------------------------------------

SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT;
SET NAMES utf8mb4;

-- ------------------------------------------------------
-- Table structure for table `user`
-- ------------------------------------------------------

DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `uid` int(11) NOT NULL AUTO_INCREMENT,
  `accountid` varchar(50) NOT NULL,
  `username` varchar(100) DEFAULT NULL,
  `data` blob DEFAULT NULL,                    -- 🔑 BLOB: protobuf 数据
  `tiny_val` tinyint(4) DEFAULT NULL,
  `small_val` smallint(6) DEFAULT NULL,
  `medium_val` mediumint(9) DEFAULT NULL,
  `int_val` int(11) DEFAULT NULL,
  `big_val` bigint(20) DEFAULT NULL,
  `float_val` float DEFAULT NULL,
  `double_val` double DEFAULT NULL,
  `decimal_val` decimal(10,2) DEFAULT NULL,
  `bool_val` tinyint(1) DEFAULT NULL,
  `date_val` date DEFAULT NULL,
  `time_val` time DEFAULT NULL,
  `datetime_val` datetime DEFAULT NULL,
  `timestamp_val` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `year_val` year(4) DEFAULT NULL,
  `char_val` char(10) DEFAULT NULL,
  `varchar_val` varchar(255) DEFAULT NULL,
  `binary_val` varbinary(255) DEFAULT NULL,
  `text_val` text,
  `blob_val` blob DEFAULT NULL,                -- 🔑 BLOB: 二进制数据
  `json_val` json DEFAULT NULL,
  `enum_val` enum('A','B','C') DEFAULT 'A',
  `set_val` set('X','Y','Z') DEFAULT NULL,
  `ctime` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `mtime` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`uid`),
  UNIQUE KEY `uk_accountid` (`accountid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------
-- Dumping data for table `user`
-- ------------------------------------------------------

LOCK TABLES `user` WRITE;
/*!40000 ALTER TABLE `user` DISABLE KEYS */;

-- 测试记录 1: 完整数据（所有字段有值）
INSERT INTO `user` VALUES
(1,'test_user_001','测试用户001',_binary '\x08\x01\x96\x01\x01\x0A\x07testing',
127,32767,8388607,2147483647,9223372036854775807,
3.14,3.1415926535,12345.67,1,
'2025-01-31','12:30:45','2025-01-31 12:30:45','2025-01-31 12:30:45',2025,
'fixed','variable',
_binary '\x00\x01\x02\x03',
'测试文本',_binary '\xE8\xB6\x8B\xE6\x96\x87',
'{"key":"value","number":42}','B','X,Y',
'2025-01-31 12:30:45','2025-01-31 12:30:45');

-- 测试记录 2: 边界值（负数、最小值）
INSERT INTO `user` VALUES
(2,'test_user_002','边界值测试',NULL,
-128,-32768,-8388608,-2147483648,-9223372036854775808,
-3.14,-3.1415926535,-12345.67,0,
'1970-01-01','00:00:00','1970-01-01 00:00:00','1970-01-01 00:00:00',1970,
'min','min_negative',
NULL,'',NULL,NULL,'A','X',
'2025-01-31 12:30:45','2025-01-31 12:30:45');

-- 测试记录 3: NULL 值测试
INSERT INTO `user` (`uid`,`accountid`,`username`) VALUES
(3,'test_user_003','NULL值测试');

-- 测试记录 4: 特殊字符
INSERT INTO `user` VALUES
(4,'test_user_004','特殊字符"测试''',_binary '\x00\xFF\x0A\x0D',
NULL,NULL,NULL,NULL,NULL,
NULL,NULL,NULL,NULL,
NULL,NULL,NULL,NULL,NULL,
'special','包含"引号"和''单引''',
NULL,'包含换行\n制表符\t',
NULL,'A',NULL,
'2025-01-31 12:30:45','2025-01-31 12:30:45');

-- 测试记录 5: 大 BLOB (1KB)
INSERT INTO `user` VALUES
(5,'test_user_005','大BLOB测试',
REPEAT(_binary '\xAB\xCD\xEF', 341),  -- ~1KB
NULL,NULL,NULL,NULL,NULL,
NULL,NULL,NULL,NULL,
NULL,NULL,NULL,NULL,
REPEAT('A',10),REPEAT('B',100),
NULL,REPEAT(_binary '\x00', 500),
NULL,'C','Y',
'2025-01-31 12:30:45','2025-01-31 12:30:45');

-- 测试记录 6-10: 扩展插入格式
INSERT INTO `user` VALUES
(6,'test_user_006','扩展插入1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'A','X','2025-01-31 12:30:45','2025-01-31 12:30:45'),
(7,'test_user_007','扩展插入2',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'B','Y','2025-01-31 12:30:45','2025-01-31 12:30:45'),
(8,'test_user_008','扩展插入3',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'C','Z','2025-01-31 12:30:45','2025-01-31 12:30:45'),
(9,'test_user_009','扩展插入4',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'A',NULL,'2025-01-31 12:30:45','2025-01-31 12:30:45'),
(10,'test_user_010','扩展插入5',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'B','X,Y','2025-01-31 12:30:45','2025-01-31 12:30:45');

/*!40000 ALTER TABLE `user` ENABLE KEYS */;
UNLOCK TABLES;

-- ------------------------------------------------------
-- End of dump
-- ------------------------------------------------------
