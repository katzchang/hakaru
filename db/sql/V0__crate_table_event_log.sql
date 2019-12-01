CREATE TABLE `eventlog` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `at` datetime DEFAULT NULL,
  `name` varchar(255) NOT NULL,
  `value` int unsigned,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4;
