const API_URL = window.__AUTH_API_URL__ || "/query_auth";
export async function graphql(query, variables = {}) {
  const response = await fetch(API_URL, { method: "POST", credentials: "include", headers: { "Content-Type": "application/json", Accept: "application/json" }, body: JSON.stringify({ query, variables }) });
  let payload; try { payload = await response.json(); } catch { throw new Error("El servidor devolvió una respuesta inválida."); }
  if (!response.ok) throw new Error(payload?.message || `Error HTTP ${response.status}`);
  if (payload.errors?.length) throw new Error(payload.errors.map((item) => item.message).join(" "));
  return payload.data;
}
export const queries = {
  login: `mutation Login($input: NewLogin2!) { login_v2(input: $input) { session_key session_time me { usuario { id nombres apellido1 apellido2 username correo foto_url estado } menus { id label path icon color grupo orden padre_id } roles { unidad { id nombre } rol { id nombre } } permisos_sueltos { metodo nombre grupo } } } }`,
  me: `query Me($input: InputMe!) { me(input: $input) { usuario { id nombres apellido1 apellido2 documento celular correo sexo direccion username estado last_login foto_url latitud longitud conexiones } menus { id label path icon color grupo orden padre_id } roles { unidad { id nombre } rol { id nombre } } permisos_sueltos { metodo nombre grupo } } }`,
  dashboard: `query Dashboard { reporte1 { nombre valor } reporte2 { fecha fecha_str valor } usuarios_conectados { id nombres apellido1 username conexiones } }`,
  users: `query Users($query: QueryUsuarios!) { usuarios2(query: $query) { paginacion { pagina size paginas registros } datos { id nombres apellido1 apellido2 username correo estado fecha_registro last_login conexiones } } }`,
  user: `query User($id: ID!) { usuario_by_id(id: $id) { id nombres apellido1 apellido2 documento celular correo sexo direccion username estado last_login foto_url latitud longitud conexiones roles { rol_id unidad_id rol_name unidad_name } permisos_sueltos { metodo nombre grupo } menus_sueltos { id label path } } }`,
  roles: `query Roles($q: QueryRoles!) { roles2(q: $q) { paginacion { pagina size paginas registros } datos { id nombre descripcion jerarquia menus permisos usuarios } } }`,
  role: `query Role($id: ID!) { rol_by_id(id: $id) { id nombre descripcion jerarquia permisos { metodo nombre descripcion grupo } menus { id label path } } }`,
  permissions: `query Permissions { permisos { metodo nombre descripcion grupo } }`,
  menus: `query Menus { menus { id label path icon color grupo orden padre_id } }`,
  units: `query Units($q: QueryUnidades!) { unidades2(q: $q) { paginacion { pagina size paginas registros } datos { id nombre descripcion latitud longitud orden fecha_registro } } }`,
  notices: `query Notices { notificaciones { id mensaje creado_por_id desde hasta fecha_registro } }`,
  tickets: `query Tickets($q: QueryTickets!) { all_tickets(q: $q) { id usuario_id usuario problema estado fecha_registro respuesta soporte soporte_id respondido } }`,
  myTickets: `query MyTickets { mis_tickets { id usuario_id usuario problema estado fecha_registro respuesta soporte soporte_id respondido } }`,
  chats: `query Chats($id: ID!) { chats_by_user(user_id: $id) { id tipo nombre foto_url no_leidos ultimo_mensaje created_by fecha_registro fecha_update miembros { usuario_id is_admin } } }`,
  messages: `query Messages($id: ID!) { chat_mensajes(conversacion_id: $id) { id conversacion_id sender_id tipo texto created_at edited_at deleted_at } }`,
};
export const mutations = {
  createUser: `mutation CreateUser($input: NewUsuario!) { create_usuario(input: $input) { id nombres apellido1 username estado } }`, updateUser: `mutation UpdateUser($input: UpdateUsuario!) { update_usuario(input: $input) { id nombres apellido1 username estado } }`,
  createRole: `mutation CreateRole($input: NewRol!) { create_rol(input: $input) { id nombre descripcion jerarquia } }`, updateRole: `mutation UpdateRole($input: UpdateRol!) { update_rol(input: $input) { id nombre descripcion jerarquia } }`,
  createUnit: `mutation CreateUnit($input: NewUnidad!) { create_unidad(input: $input) { id nombre descripcion orden } }`, updateUnit: `mutation UpdateUnit($input: UpdUnidad!) { update_unidad(input: $input) { id nombre descripcion orden } }`,
  createNotice: `mutation CreateNotice($input: NewNotificacion!) { crear_notificacion(input: $input) { id mensaje desde hasta } }`, updateNotice: `mutation UpdateNotice($input: UpdNotificacion!) { update_notificacion(input: $input) { id mensaje desde hasta } }`,
  createTicket: `mutation CreateTicket($input: NewTicket!) { create_ticket(input: $input) { id problema estado fecha_registro } }`, answerTicket: `mutation AnswerTicket($input: NewTicketRespuesta!) { update_ticket(input: $input) { id estado respuesta } }`, closeTicket: `mutation CloseTicket($id: ID!) { cerrar_ticket(id: $id) { id estado } }`,
  sendMessage: `mutation SendMessage($input: ChatEnviarMensajeInput!) { chat_enviar_mensaje(input: $input) { id texto created_at } }`, updateProfile: `mutation UpdateProfile($input: UpdatePerfil!) { update_perfil(input: $input) { id nombres apellido1 username correo } }`,
};
