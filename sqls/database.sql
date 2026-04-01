create table `rbac_usuarios`(
    `id` integer unsigned auto_increment not null primary key,
    `nombres` varchar(60) not null,
    `apellido1` varchar(30) not null,
    `apellido2` varchar(30),
    `documento` varchar(30),
    `celular` varchar(20),
    `correo` varchar(120),
    `sexo` enum('M','F'),
    `direccion` varchar(100),
    `estado` tinyint not null default 1,
    `username` varchar(30) unique not null,
    `password` varchar(80) not null, -- hash
    `last_login` datetime,
    `oauth_id` varchar(80),
    `foto_url` varchar(90),
    `ubicacion` point, 
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    `fecha_update` datetime not null default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

create table `rbac_roles`(
    `id` smallint unsigned auto_increment not null primary key,
    `nombre` varchar(50) not null unique,
    `descripcion` varchar(100),
    `jerarquia` tinyint not null default 0,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP
);

-- sedes, departamentos, ubicaciones segun nomenclatura
create table `rbac_unidades`(
    `id` smallint unsigned auto_increment not null primary key,
    `nombre` varchar(70) not null unique,
    `descripcion` varchar(100),
    `ubicacion` point, 
    `orden` tinyint not null default 0,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP
);

-- para permisos no hay crud, se agregan segun se crean las funcionaes 
-- en el propio codigo fuente 
create table `rbac_permisos`(
    -- en golang hay funciones(query,mutation), cada funcion representa un permiso, el nombre de esa funcion 
    -- es el valor de metodo, (ej createPersona)
    `metodo` varchar(50) not null primary key,
    -- el nombre es el mismo que metodo, pero para que un humano lo lea, (ej: crear persona)
    `nombre` varchar(50) not null,
    `descripcion` varchar(200),
    `grupo` varchar(30) not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP
);

create table `rbac_rol_permiso`(
    `rol_id` smallint unsigned not null,
    `metodo` varchar(50) not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`rol_id`) references `rbac_roles`(`id`),
    foreign key(`metodo`) references `rbac_permisos`(`metodo`),
    primary key(`rol_id`,`metodo`)
);

create table `rbac_rol_usuario_unidades`(
    `rol_id` smallint unsigned not null,
    `usuario_id` integer unsigned not null,
    `unidad_id` smallint unsigned not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`rol_id`) references `rbac_roles`(`id`),
    foreign key(`usuario_id`) references `rbac_usuarios`(`id`),
    foreign key(`unidad_id`) references `rbac_unidades`(`id`),
    primary key(`rol_id`,`usuario_id`,`unidad_id`)
);

create table `rbac_usuario_permiso`(
    `usuario_id` integer unsigned not null,
    `metodo` varchar(50) not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`usuario_id`) references `rbac_usuarios`(`id`),
    foreign key(`metodo`) references `rbac_permisos`(`metodo`),
    primary key(`usuario_id`,`metodo`)
);

create table `rbac_menus`(
    `id` tinyint unsigned auto_increment not null primary key,
    `label` varchar(40) not null unique,
    `path` varchar(40) not null,
    `icon` varchar(40) not null,
    `color` varchar(40) not null,
    `grupo` tinyint unsigned not null default 1,
    `orden` tinyint unsigned not null default 1,
    `padre_id` tinyint unsigned null,
    foreign key(`padre_id`) references `rbac_menus`(`id`)
);

create table `rbac_menus_usuario`(
    `id` smallint unsigned auto_increment not null primary key,
    `usuario_id` integer unsigned not null,
    `menu_id` tinyint unsigned not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`usuario_id`) references `rbac_usuarios`(`id`),
    foreign key(`menu_id`) references `rbac_menus`(`id`)
);

create table `rbac_rol_menus`(
    `id` smallint unsigned auto_increment not null primary key,
    `rol_id` smallint unsigned not null,
    `menu_id` tinyint unsigned not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`rol_id`) references `rbac_roles`(`id`),
    foreign key(`menu_id`) references `rbac_menus`(`id`)
);

create table `rbac_session_keys`(
    `id` integer unsigned auto_increment not null primary key,
    `usuario_id` integer unsigned not null,
    `key` varchar(80) not null unique,
    `apikey` varchar(255) not null,
    `expire` datetime not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`usuario_id`) references `rbac_usuarios`(`id`)
);

create table `rbac_notificaciones`(
    `id` integer unsigned auto_increment not null primary key,
    `mensaje` text not null, -- formato html
    `creado_por_id` integer unsigned not null,
    `desde` datetime not null,
    `hasta` datetime not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`creado_por_id`) references `rbac_usuarios`(`id`)
);

create table `rbac_tickets`(
    `id` integer unsigned auto_increment not null primary key,
    `usuario_id` integer unsigned not null,
    `problema` text not null,
    `estado` enum('pendiente','cliente','soporte','cerrado') not null default 'pendiente',
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`usuario_id`) references `rbac_usuarios`(`id`)
);

create table `rbac_tickets_respuestas`(
    `id` integer unsigned auto_increment not null primary key,
    `tickets_id` integer unsigned not null,
    `usuario_id` integer unsigned not null,
    `respuesta` text not null,
    `fecha_registro` datetime not null default CURRENT_TIMESTAMP,
    foreign key(`usuario_id`) references `rbac_usuarios`(`id`),
    foreign key(`tickets_id`) references `rbac_tickets`(`id`)
);


-- indice para optimizar la busqueda 
-- CREATE INDEX idx_username ON usuarios (username);
CREATE INDEX idx_rol_usuario_unidades ON rbac_rol_usuario_unidades(usuario_id, rol_id, unidad_id);
CREATE INDEX idx_rol_permiso ON rbac_rol_permiso(rol_id, metodo);
ALTER TABLE rbac_tickets AUTO_INCREMENT = 100;

insert into `rbac_unidades`(`nombre`) values('principal');
insert into `rbac_unidades`(`nombre`) values('secundaria');

-- los id de menus estan reservados hasta el 10
-- en tu app debes crearlas desde el 10 o superior
insert into `rbac_menus`(`id`,`label`,`path`,`icon`,`grupo`,`color`,`orden`,`padre_id`) 
values 
(1,'Sistema','/','manage_accounts',1,'primary',100, null),
(2,'Usuarios','/usuarios','group',1,'primary',1, 1),
(3,'Roles','/roles','local_movies',1,'primary',2, 1),
(4,'Unidades','/unidades','home',1,'primary',3, 1),
(5,'Notificaciones','/avisos','campaign',1,'primary',4, 1),
(6,'Tickets','/tickets','confirmation_number',1,'primary',5, 1);


INSERT INTO `rbac_permisos` (`metodo`, `nombre`, `grupo`,`descripcion`)
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

insert into `rbac_roles`(`id`,`nombre`,`descripcion`,`jerarquia`)
values 
(1,'Administrador','Total acceso',0),
(2,'Invitado','Bienvenido',1),
(3,'Externo','Bienvenido',2);

insert into `rbac_rol_menus`(`rol_id`,`menu_id`)
values (1,1), (1,2), (1,3), (1,4), (1,5),(1,6), (2,6), (3,6);

insert into `rbac_usuarios`(`nombres`,`apellido1`,`username`,`password`)
values ('admin','','admin',SHA2('admin', 256));

insert into `rbac_rol_usuario_unidades`(`rol_id`,`usuario_id`,`unidad_id`)
values (1,1,1), (2,1,2);

insert into `rbac_rol_permiso`(`rol_id`,`metodo`)
values (1,'roles'),(1,'rol_by_id'),(1,'update_rol'),(1,'permisos'),(1,'menus');

-- 
-- ejemplo de menus anidados: 
-- 
-- insert into `rbac_menus`(`id`,`label`,`path`,`icon`,`grupo`,`color`,`orden`,`padre_id`) 
-- values 
-- (6,'Menu anidado','/','group',2,'primary',1,null),
-- (7,'Rolesx','/roles','local_movies',2,'primary',2, 6),
-- (8,'Unidadesx','/unidades','home',2,'primary',3, 6),
-- (9,'Notificacionesx','/avisos','campaign',2,'primary',4, 6);