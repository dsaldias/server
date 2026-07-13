-- ======================================================================================
-- NEW feature chats
-- Representa un chat. Puede ser privado o grupal.
create table `rbac_conversaciones`(
    `id` integer unsigned auto_increment not null primary key,
    `tipo` enum('privado','grupo') not null default 'privado',
    `nombre` varchar(50) not null,
    `foto_url` varchar(90),
    `created_by` integer unsigned not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    `fecha_update` datetime not null default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    foreign key(`created_by`) references `rbac_usuarios`(`id`)
);

-- La tabla principal.
create table `rbac_mensajes`(
    `id` integer unsigned auto_increment not null primary key,
    `conversacion_id` integer unsigned not null,
    `sender_id` integer unsigned not null,
    `tipo` enum('text','image','video','audio','file','location','system') not null,
    `texto` varchar(256) not null,
    `created_at` datetime not null default CURRENT_TIMESTAMP,
    `edited_at` datetime,
    `deleted_at`datetime,
    foreign key(`conversacion_id`) references `rbac_conversaciones`(`id`),
    foreign key(`sender_id`) references `rbac_usuarios`(`id`)
);

-- Quién pertenece a cada conversación.
create table `rbac_conversaciones_miembros`(
    `id` integer unsigned auto_increment not null primary key,
    `conversacion_id` integer unsigned not null,
    `usuario_id` integer unsigned not null,
    `is_admin` tinyint(1) not null default 0,
    `joined_at` datetime not null default CURRENT_TIMESTAMP,
    `left_at` datetime,
    `last_read_message_id` integer unsigned,
    unique key `uk_conversacion_usuario` (`conversacion_id`, `usuario_id`),
    foreign key(`conversacion_id`) references `rbac_conversaciones`(`id`),
    foreign key(`usuario_id`) references `rbac_usuarios`(`id`),
    foreign key(`last_read_message_id`) references `rbac_mensajes`(`id`)
);

-- Porque tarde o temprano alguien quiere mandar un PDF de 200 MB diciendo es livianito.
create table `rbac_mensaje_files`(
    `id` integer unsigned auto_increment not null primary key,
    `message_id` integer unsigned not null,
    `filename` varchar(50) not null,
    `file_path` varchar(90) not null,
    `created_at` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`message_id`) references `rbac_mensajes`(`id`)
);

-- Para saber quién recibió y leyó.
create table `rbac_mensaje_status`(
    `id` integer unsigned auto_increment not null primary key,
    `message_id` integer unsigned not null,
    `usuario_id` integer unsigned not null,
    `status` enum('sent','delivered','read') not null,
    `updated_at` datetime not null default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    unique key `uk_mensaje_usuario` (`message_id`, `usuario_id`),
    foreign key(`message_id`) references `rbac_mensajes`(`id`),
    foreign key(`usuario_id`) references `rbac_usuarios`(`id`)
);

-- END feature chats