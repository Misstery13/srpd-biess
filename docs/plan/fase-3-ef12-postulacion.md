# Fase 3 — EF-12 Formulario de postulación

**Fecha:** domingo 1 – lunes 2 · **Hito H3:** el postulante `1710034065` llena todas las secciones, guarda, recarga sin perder datos y envía; `1754321980` (colombiano) queda bloqueado.

## D — Datos
- [ ] `PostulacionRepository.Guardar(tx)`: **una transacción** sobre `postulaciones`, `domicilios`, `familiares_iess`, `formacion_academica`, `experiencia_laboral`, `publicaciones`, `capacitaciones` e `inhabilidades`, con `ROLLBACK` ante cualquier error (P4).
- [ ] `Obtener(postulante_id)` con todas las secciones.
- [ ] Código único `BIESS-2026-000001` (P2).
- [ ] Prueba: forzar un error en la última inserción y comprobar que no queda nada guardado.

## B — Backend
- [ ] `GET /postulante/me`: 403 si no es ecuatoriano (V5).
- [ ] `validar-cargo`: afiliado / jubilado / empleador con los mensajes exactos de 1.4.x.
- [ ] Validaciones de servidor: campos obligatorios (*), fechas (V7), máx. 1000 caracteres (V8), estado `ENVIADA` no editable (P3).
- [ ] Subida de PDF: tipo real y tamaños (A1–A3).
- [ ] `enviar`: exige `acepta_terminos`, cambia el estado y simula el correo (I2).

## F — Frontend
- [ ] Formulario por secciones (pestañas o pasos): Personal (solo lectura, P6, con pareja si está casado/a o en unión de hecho) · Domicilio · Información adicional (familiares) · Cargo · Formación · Experiencia · Publicaciones · Capacitación · Inhabilidades.
- [ ] Listas con Agregar / Editar / Eliminar en familiares, formación, experiencia, publicaciones y capacitación.
- [ ] Calculadora de experiencia (años, meses, días).
- [ ] Botón "Guardar progreso" en cada sección; "Guardar postulación" muestra los 5 textos de consentimiento; "Enviar postulación".
- [ ] Mensaje de cierre si el postulante no es ecuatoriano o no cumple el cargo.

## Criterio de aceptación
- [ ] Prueba de ROLLBACK documentada (captura o consulta SQL).
- [ ] Los 4 usuarios de prueba se comportan según el README.
