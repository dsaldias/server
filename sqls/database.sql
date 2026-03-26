create table `usuarios`(
    `id` integer unsigned auto_increment not null primary key,
    `nombres` varchar(60) not null,
    `apellido1` varchar(30) not null,
    `apellido2` varchar(30),
    `documento` varchar(30),
    `celular` varchar(20),
    `correo` varchar(100),
    `sexo` enum('M','F'),
    `direccion` varchar(100),
    `estado` tinyint(1) not null default 1,
    `username` varchar(30) unique not null,
    `password` varchar(64) not null, -- hash
    `last_login` datetime,
    `oauth_id` varchar(80),
    `foto_url` varchar(90),
    `ubicacion` point, 
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    `fecha_update` datetime not null default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

create table `roles`(
    `id` smallint unsigned auto_increment not null primary key,
    `nombre` varchar(50) not null unique,
    `descripcion` varchar(100),
    `jerarquia` tinyint(1) not null default 0,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP
);

-- sedes, departamentos, ubicaciones segun nomenclatura
create table `unidades`(
    `id` smallint unsigned auto_increment not null primary key,
    `nombre` varchar(70) not null unique,
    `descripcion` varchar(100),
    `ubicacion` point, 
    `orden` tinyint(1) not null default 0,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP
);

-- para permisos no hay crud, se agregan segun se crean las funcionaes 
-- en el propio codigo fuente 
create table `permisos`(
    -- en golang hay funciones(query,mutation), cada funcion representa un permiso, el nombre de esa funcion 
    -- es el valor de metodo, (ej createPersona)
    `metodo` varchar(50) not null primary key,
    -- el nombre es el mismo que metodo, pero para que un humano lo lea, (ej: crear persona)
    `nombre` varchar(50) not null,
    `descripcion` varchar(200),
    `grupo` varchar(30) not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP
);

create table `rol_permiso`(
    `rol_id` smallint unsigned not null,
    `metodo` varchar(50) not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`rol_id`) references `roles`(`id`),
    foreign key(`metodo`) references `permisos`(`metodo`),
    primary key(`rol_id`,`metodo`)
);

create table `rol_usuario_unidades`(
    `rol_id` smallint unsigned not null,
    `usuario_id` integer unsigned not null,
    `unidad_id` smallint unsigned not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`rol_id`) references `roles`(`id`),
    foreign key(`usuario_id`) references `usuarios`(`id`),
    foreign key(`unidad_id`) references `unidades`(`id`),
    primary key(`rol_id`,`usuario_id`,`unidad_id`)
);

create table `usuario_permiso`(
    `usuario_id` integer unsigned not null,
    `metodo` varchar(50) not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`usuario_id`) references `usuarios`(`id`),
    foreign key(`metodo`) references `permisos`(`metodo`),
    primary key(`usuario_id`,`metodo`)
);

create table `menus`(
    `id` tinyint unsigned auto_increment not null primary key,
    `label` varchar(40) not null unique,
    `path` varchar(40) not null,
    `icon` varchar(40) not null,
    `color` varchar(40) not null,
    `grupo` tinyint(1) unsigned not null default 1,
    `orden` tinyint(1) unsigned not null default 1
);

create table `menus_usuario`(
    `id` tinyint unsigned auto_increment not null primary key,
    `usuario_id` integer unsigned not null,
    `menu_id` tinyint unsigned not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`usuario_id`) references `usuarios`(`id`),
    foreign key(`menu_id`) references `menus`(`id`)
);

create table `rol_menus`(
    `id` tinyint unsigned auto_increment not null primary key,
    `rol_id` smallint unsigned not null,
    `menu_id` tinyint unsigned not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`rol_id`) references `roles`(`id`),
    foreign key(`menu_id`) references `menus`(`id`)
);

create table `session_keys`(
    `id` integer unsigned auto_increment not null primary key,
    `usuario_id` integer unsigned not null,
    `key` varchar(80) not null unique,
    `apikey` varchar(255) not null,
    `expire` datetime not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`usuario_id`) references `usuarios`(`id`)
);

create table `notificaciones`(
    `id` integer unsigned auto_increment not null primary key,
    `mensaje` text not null, -- formato html
    `creado_por_id` integer unsigned not null,
    `desde` datetime not null,
    `hasta` datetime not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`creado_por_id`) references `usuarios`(`id`)
);

create table `tickets`(
    `id` integer unsigned auto_increment not null primary key,
    `usuario_id` integer unsigned not null,
    `problema` text not null,
    `estado` enum('pendiente','cliente','soporte','cerrado') not null default 'pendiente',
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`usuario_id`) references `usuarios`(`id`)
);

create table `tickets_respuestas`(
    `id` integer unsigned auto_increment not null primary key,
    `tickets_id` integer unsigned not null,
    `usuario_id` integer unsigned not null,
    `respuesta` text not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`usuario_id`) references `usuarios`(`id`),
    foreign key(`tickets_id`) references `tickets`(`id`)
);


-- indice para optimizar la busqueda 
CREATE INDEX idx_username ON usuarios (username);
CREATE INDEX idx_rol_usuario_unidades ON rol_usuario_unidades(usuario_id, rol_id, unidad_id);
CREATE INDEX idx_rol_permiso ON rol_permiso(rol_id, metodo);
ALTER TABLE tickets AUTO_INCREMENT = 100;

insert into `unidades`(`nombre`) values('principal');
insert into `unidades`(`nombre`) values('secundaria');

insert into `menus`(`id`,`label`,`path`,`icon`,`grupo`,`color`,`orden`) 
values 
(1,'Usuarios','/usuarios','group',1,'primary',1),
(2,'Roles','/roles','local_movies',1,'primary',2),
(3,'Unidades','/unidades','home',1,'primary',3),
(4,'Notificaciones','/avisos','campaign',1,'primary',4),
(5,'Tickets','/tickets','confirmation_number',1,'primary',5);


INSERT INTO `permisos` (`metodo`, `nombre`, `grupo`,`descripcion`)
VALUES
('roles', 'roles', 'roles','Listar roles'),
('permisos', 'permisos', 'permisos','Listar permisos'),
('usuarios', 'usuarios', 'usuarios','Listar usuarios'),
('usuario_by_id', 'usuario by id', 'usuarios','Lista un usuario'),
('rol_by_id', 'rol by id', 'roles','Listar un rol'),
('menus', 'menus', 'menus','Listar menus'),
('unidades', 'unidades', 'unidades','Listar unidades'),
('create_rol', 'create rol', 'roles','Crear rol'),
('update_rol', 'update rol', 'roles','Actualizar rol'),
('create_usuario', 'create usuario', 'usuarios','Crear usuario'),
('update_usuario', 'update usuario', 'usuarios','Actualizar usuario'),
('update_perfil', 'update perfil', 'usuarios','Actualizar perfil'),
('create_unidad', 'create unidad', 'unidades','Crear unidad'),
('update_unidad', 'update unidad', 'unidades','ctualizar unidad'),
('crear_notificacion', 'crear notificacion', 'notificacion','crea un nuevo aviso'),
('update_notificacion', 'actualizar notificacion', 'notificacion','actualiza el aviso'),
('ver_ticket', 'ver tickets', 'tickets','ver detalles del ticket'),
('mis_tickets', 'listar mis tickets', 'tickets','tickets del usuario'),
('all_tickets', 'listar todos los tickets', 'tickets','todos los tickets'),
('create_ticket', 'crear tickets', 'tickets','crear tickets'),
('update_ticket', 'responder tickets', 'tickets','responder tickets'),
('cerrar_ticket', 'cerrar tickets', 'tickets','cerrar tickets');

insert into `roles`(`id`,`nombre`,`descripcion`,`jerarquia`)
values 
(1,'Administrador','Total acceso',0),
(2,'Invitado','Bienvenido',1),
(3,'Externo','Bienvenido',2);

insert into `rol_menus`(`rol_id`,`menu_id`)
values (1,1), (1,2), (1,3), (1,4), (1,5), (2,1);

insert into `usuarios`(`nombres`,`apellido1`,`username`,`password`)
values ('admin','','admin',SHA2('admin', 256));

insert into `rol_usuario_unidades`(`rol_id`,`usuario_id`,`unidad_id`)
values (1,1,1), (2,1,2);

insert into `rol_permiso`(`rol_id`,`metodo`)
values (1,'roles'),(1,'rol_by_id'),(1,'update_rol'),(1,'permisos'),(1,'menus');

