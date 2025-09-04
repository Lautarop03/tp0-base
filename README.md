# TP0: Docker + Comunicaciones + Concurrencia

En el presente repositorio se provee un esqueleto básico de cliente/servidor, en donde todas las dependencias del mismo se encuentran encapsuladas en containers. Los alumnos deberán resolver una guía de ejercicios incrementales, teniendo en cuenta las condiciones de entrega descritas al final de este enunciado.

 El cliente (Golang) y el servidor (Python) fueron desarrollados en diferentes lenguajes simplemente para mostrar cómo dos lenguajes de programación pueden convivir en el mismo proyecto con la ayuda de containers, en este caso utilizando [Docker Compose](https://docs.docker.com/compose/).

## Instrucciones de uso
El repositorio cuenta con un **Makefile** que incluye distintos comandos en forma de targets. Los targets se ejecutan mediante la invocación de:  **make \<target\>**. Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de depuración.

Los targets disponibles son:

| target  | accion  |
|---|---|
|  `docker-compose-up`  | Inicializa el ambiente de desarrollo. Construye las imágenes del cliente y el servidor, inicializa los recursos a utilizar (volúmenes, redes, etc) e inicia los propios containers. |
| `docker-compose-down`  | Ejecuta `docker-compose stop` para detener los containers asociados al compose y luego  `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene de versiones de desarrollo y recursos sin liberar. |
|  `docker-compose-logs` | Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose. |
| `docker-image`  | Construye las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para probar nuevos cambios en las imágenes antes de arrancar el proyecto. |
| `build` | Compila la aplicación cliente para ejecución en el _host_ en lugar de en Docker. De este modo la compilación es mucho más veloz, pero requiere contar con todo el entorno de Golang y Python instalados en la máquina _host_. |

### Servidor

Se trata de un "echo server", en donde los mensajes recibidos por el cliente se responden inmediatamente y sin alterar. 

Se ejecutan en bucle las siguientes etapas:

1. Servidor acepta una nueva conexión.
2. Servidor recibe mensaje del cliente y procede a responder el mismo.
3. Servidor desconecta al cliente.
4. Servidor retorna al paso 1.


### Cliente
 se conecta reiteradas veces al servidor y envía mensajes de la siguiente forma:
 
1. Cliente se conecta al servidor.
2. Cliente genera mensaje incremental.
3. Cliente envía mensaje al servidor y espera mensaje de respuesta.
4. Servidor responde al mensaje.
5. Servidor desconecta al cliente.
6. Cliente verifica si aún debe enviar un mensaje y si es así, vuelve al paso 2.

### Ejemplo

Al ejecutar el comando `make docker-compose-up`  y luego  `make docker-compose-logs`, se observan los siguientes logs:

```
client1  | 2024-08-21 22:11:15 INFO     action: config | result: success | client_id: 1 | server_address: server:12345 | loop_amount: 5 | loop_period: 5s | log_level: DEBUG
client1  | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:14 DEBUG    action: config | result: success | port: 12345 | listen_backlog: 5 | logging_level: DEBUG
server   | 2024-08-21 22:11:14 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°3
client1  | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°3
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°5
client1  | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°5
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:40 INFO     action: loop_finished | result: success | client_id: 1
client1 exited with code 0
```


## Parte 1: Introducción a Docker
En esta primera parte del trabajo práctico se plantean una serie de ejercicios que sirven para introducir las herramientas básicas de Docker que se utilizarán a lo largo de la materia. El entendimiento de las mismas será crucial para el desarrollo de los próximos TPs.

### Ejercicio N°1:
Definir un script de bash `generar-compose.sh` que permita crear una definición de Docker Compose con una cantidad configurable de clientes.  El nombre de los containers deberá seguir el formato propuesto: client1, client2, client3, etc. 

El script deberá ubicarse en la raíz del proyecto y recibirá por parámetro el nombre del archivo de salida y la cantidad de clientes esperados:

`./generar-compose.sh docker-compose-dev.yaml 5`

Considerar que en el contenido del script pueden invocar un subscript de Go o Python:

```
#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"
python3 mi-generador.py $1 $2
```

En el archivo de Docker Compose de salida se pueden definir volúmenes, variables de entorno y redes con libertad, pero recordar actualizar este script cuando se modifiquen tales definiciones en los sucesivos ejercicios.

### Ejercicio N°2:
Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida por fuera de la imagen (hint: `docker volumes`).


### Ejercicio N°3:
Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. Dado que el servidor es un echo server, se debe enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado.

En caso de que la validación sea exitosa imprimir: `action: test_echo_server | result: success`, de lo contrario imprimir:`action: test_echo_server | result: fail`.

El script deberá ubicarse en la raíz del proyecto. Netcat no debe ser instalado en la máquina _host_ y no se pueden exponer puertos del servidor para realizar la comunicación (hint: `docker network`). `


### Ejercicio N°4:
Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos, sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. Loguear mensajes en el cierre de cada recurso (hint: Verificar que hace el flag `-t` utilizado en el comando `docker compose down`).

## Parte 2: Repaso de Comunicaciones

Las secciones de repaso del trabajo práctico plantean un caso de uso denominado **Lotería Nacional**. Para la resolución de las mismas deberá utilizarse como base el código fuente provisto en la primera parte, con las modificaciones agregadas en el ejercicio 4.

### Ejercicio N°5:
Modificar la lógica de negocio tanto de los clientes como del servidor para nuestro nuevo caso de uso.

#### Cliente
Emulará a una _agencia de quiniela_ que participa del proyecto. Existen 5 agencias. Deberán recibir como variables de entorno los campos que representan la apuesta de una persona: nombre, apellido, DNI, nacimiento, numero apostado (en adelante 'número'). Ej.: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574` respectivamente.

Los campos deben enviarse al servidor para dejar registro de la apuesta. Al recibir la confirmación del servidor se debe imprimir por log: `action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}`.



#### Servidor
Emulará a la _central de Lotería Nacional_. Deberá recibir los campos de la cada apuesta desde los clientes y almacenar la información mediante la función `store_bet(...)` para control futuro de ganadores. La función `store_bet(...)` es provista por la cátedra y no podrá ser modificada por el alumno.
Al persistir se debe imprimir por log: `action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

#### Comunicación:
Se deberá implementar un módulo de comunicación entre el cliente y el servidor donde se maneje el envío y la recepción de los paquetes, el cual se espera que contemple:
* Definición de un protocolo para el envío de los mensajes.
* Serialización de los datos.
* Correcta separación de responsabilidades entre modelo de dominio y capa de comunicación.
* Correcto empleo de sockets, incluyendo manejo de errores y evitando los fenómenos conocidos como [_short read y short write_](https://cs61.seas.harvard.edu/site/2018/FileDescriptors/).


### Ejercicio N°6:
Modificar los clientes para que envíen varias apuestas a la vez (modalidad conocida como procesamiento por _chunks_ o _batchs_). 
Los _batchs_ permiten que el cliente registre varias apuestas en una misma consulta, acortando tiempos de transmisión y procesamiento.

La información de cada agencia será simulada por la ingesta de su archivo numerado correspondiente, provisto por la cátedra dentro de `.data/datasets.zip`.
Los archivos deberán ser inyectados en los containers correspondientes y persistido por fuera de la imagen (hint: `docker volumes`), manteniendo la convencion de que el cliente N utilizara el archivo de apuestas `.data/agency-{N}.csv` .

En el servidor, si todas las apuestas del *batch* fueron procesadas correctamente, imprimir por log: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de detectar un error con alguna de las apuestas, debe responder con un código de error a elección e imprimir: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.

La cantidad máxima de apuestas dentro de cada _batch_ debe ser configurable desde config.yaml. Respetar la clave `batch: maxAmount`, pero modificar el valor por defecto de modo tal que los paquetes no excedan los 8kB. 

Por su parte, el servidor deberá responder con éxito solamente si todas las apuestas del _batch_ fueron procesadas correctamente.

### Ejercicio N°7:

Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.

El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.

Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.

## Parte 3: Repaso de Concurrencia
En este ejercicio es importante considerar los mecanismos de sincronización a utilizar para el correcto funcionamiento de la persistencia.

### Ejercicio N°8:

Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo. En caso de que el alumno implemente el servidor en Python utilizando _multithreading_,  deberán tenerse en cuenta las [limitaciones propias del lenguaje](https://wiki.python.org/moin/GlobalInterpreterLock).

## Condiciones de Entrega
Se espera que los alumnos realicen un _fork_ del presente repositorio para el desarrollo de los ejercicios y que aprovechen el esqueleto provisto tanto (o tan poco) como consideren necesario.

Cada ejercicio deberá resolverse en una rama independiente con nombres siguiendo el formato `ej${Nro de ejercicio}`. Se permite agregar commits en cualquier órden, así como crear una rama a partir de otra, pero al momento de la entrega deberán existir 8 ramas llamadas: ej1, ej2, ..., ej7, ej8.
 (hint: verificar listado de ramas y últimos commits con `git ls-remote`)

Se espera que se redacte una sección del README en donde se indique cómo ejecutar cada ejercicio y se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado (Parte 2) y los mecanismos de sincronización utilizados (Parte 3).

Se proveen [pruebas automáticas](https://github.com/7574-sistemas-distribuidos/tp0-tests) de caja negra. Se exige que la resolución de los ejercicios pase tales pruebas, o en su defecto que las discrepancias sean justificadas y discutidas con los docentes antes del día de la entrega. El incumplimiento de las pruebas es condición de desaprobación, pero su cumplimiento no es suficiente para la aprobación. Respetar las entradas de log planteadas en los ejercicios, pues son las que se chequean en cada uno de los tests.

La corrección personal tendrá en cuenta la calidad del código entregado y casos de error posibles, se manifiesten o no durante la ejecución del trabajo práctico. Se pide a los alumnos leer atentamente y **tener en cuenta** los criterios de corrección informados  [en el campus](https://campusgrado.fi.uba.ar/mod/page/view.php?id=73393).

## Decisiones de diseño
### Ej1
Se realizo un script de bash `generar-compose.sh`, decidi hacerlo en Bash sin apoyarme en python.
Al tratarse de un compose simple, esta solucion es suficiente. En otros escenarios mas complejos optaria por Python siendo este mas conveniente.

### Ej2
En docker compose se agrega un volumen para cada contenedor, tanto del cliente como del servidor. El volumen se define de la siguiente forma:
``` 
volumes:
    - ./server/config.ini:/config.ini
```
Con esta configuracion, el archivo `config` en el directorio host se monta dentro del contenedor en la ruta `/config`

### Ej3
En el script `validar-echo-server` se usa la imagen `busybox` que contiene `netcat` para conectarse al servidor y comprobar el correcto funcionamiento del mismo.
```
docker run --rm --network $NETWORK busybox sh -c "echo $MESSAGE | nc -w 3 $SERVER_IP $SERVER_PORT"
```
Flags:
`--rm` hace que sea un contenedor temporal y se elimine al finalizar la ejecucion.

`--network $NETWORK` conecta el contenedor a la red interna de docker donde se encuentra el server.

`busybox sh -c "..."` ejecuta un shell dentro de la imagen `busybox`

### Ej4
El flag `-t` en `docker compose down` define cuantos segundos docker espera a que el contenedor se detenga de forma gracegul despues de enviarle la señal `SIGTERM`

En el server se agrego un handler que recibe la señal `SIGTERM` y asi detener de forma graceful.
```
signal.signal(signal.SIGTERM, make_graceful_shutdown(server))
```
Cuando se reciba la señal, se ejecuta la funcion `make_graceful_shutdown(server)` esta es una clousure para poder tener referencia al server, devuelve otra funcion que, al ejecutarse, llama a `server.stop()` y esta ultima se encarga de cerrar los sockets del servidor y cliente.

En el cliente se agrega un canal de tamaño 1 para recibir señales
```
sigs := make(chan os.Signal, 1)
signal.Notify(sigs, syscall.SIGTERM)
done := make(chan bool, 1)
```
Las señales `SIGTERM` se enviaran por el canal sigs

Luego se lanza una goroutine que se queda esperando por la señal en sigs
```
go func() {
    <-sigs
    client.Stop()
    done <- true
}()
```
Se llama a `client.stop()` que cierra los sockets correspondientes y luego se envia `true` por el canal `done` para avisar al main que puede finalizar.

### Ej5

#### Confirmación de apuestas

Despues de recibir y guardar una apuesta, el servidor sigue funcionando como un echo server para confirmar la recepcion.

El servidor envía de vuelta la misma apuesta que recibio.

Esto permite al cliente verificar que todos los campos se recibieron correctamente.

#### Definicion del protocolo

Cada mensaje de apuesta se compone de varios campos de tipo string, correspondientes a una apuesta.

Para cada campo se sigue el formato:

`[longitud: 2 bytes uint16 big-endian]` `[string en UTF-8]`


Los 2 bytes iniciales indican la longitud exacta del string que sigue.


Campos transmitidos en orden:

`client_id (ID del cliente o agencia)`

`first_name (nombre)`

`last_name (apellido)`

`document (documento)`

`birthdate (fecha de nacimiento)`

`number (numero de apuesta)`

#### Uso correcto de sockets

*Short-read server:*
```
def read_exact(sock, n):
    buf = b''
    while len(buf) < n:
        chunk = sock.recv(n - len(buf))
        if not chunk:
            raise ConnectionError("Connection closed before receiving all bytes")
        buf += chunk
    return buf
```
Esta función asegura que se lean exactamente los bytes esperados.

*Short-write server:*

`sendall()` asegura que todos los bytes del mensaje se envíen al socket.

*Short-write client:*
```
total := 0
for total < len(data) {
	n, err := conn.Write(data[total:])
	if err != nil {
		return err
	}
	total += n
}
return nil
```

### Ej6
Se monta el archivo de data para cada cliente segun su id.
```
./.data/agency-&1.csv:/agency.csv
```
De esta manera tenemos que todos los clientes independietemente de su id, utilizan dentro de su contenedor el nombre de archivo `/agency.csv`.

#### Protocolo de batch
Como cliente se lee el archivo csv con `"encoding/csv"`, se llena el batch con apuestas hasta llegar al limite de 8kB (constante configurable dentro del archivo client.go) o `batch.maxAmount:` (en `config.yaml`).
Por cada batch se serializa y envia el mismo de la siguiente manera:

`[cantidad de apuestas enviadas en el batch: 2 bytes uint16 big-endian]`

Luego, todas las apuestas se envian seguidas, sin delimitadores. El servidor solo necesita saber cuantas apuestas recibir para procesarlas y guardarlas.

### Otras modificaciones
* Se elimino la reconexion luego de cada mensaje
* El servidor escucha una unica conexion hasta que el cliente la cierra, se modifico para este ejercicio y asi un cliente envia todas las apuestas mediante batchs y no necesita reconectarse. El cliente al finalizar cierra el socket y el servidor queda disponible escuchando nuevas conexiones.
- **Modificacion de diseño de protocolo**: 
Se realizo una actualizacion en el protocolo con respecto a los tamaños de los indicadores de longitud de los strings.
Ahora son `uint8`, en base a los csv recibidos me parece una buena opcion tener hasta un largo de 255 bytes. Esto tambien permite el envio de mayores apuestas dentro del batch.

### Ej7
- En los ejercicios anteriores el cliente enviaba mensaje de manera secuencial y el servidor esperaba estos mensajes. Para poder manejar distintos casos como se presentan en este ejercicio modifique nuevamente el protocolo de comunicacion. De esta manera no dependemos de un orden y podemos procesar los distintos mensajes que pueden llegar.
Cada mensaje comienza con un byte de opcode, que indica el tipo de mensaje y como debe recibirse el resto de la informacion.

| Opcode | Nombre | Estructura de datos | Descripcion  |
|--------|--------|---------------------|--------------|
| `0x01` | InitClient | `1 byte: length` + `N bytes clientID` | El cliente envia su ID hacia el servidor |
| `0x02` | Batch Bets | `2 bytes: length` + `N Bets` | El cliente envia al server un batch de N bets (el protocolo de envio de bets se mantiene igual) |
| `0x03` | Batch Confirmation | `1 byte: succes` | El server confirma el procesamiento del batch |
| `0x04` | Finished bets | - | El cliente confirma al server que ya envio todas las Bets | 
| `0x05` | Request Winners | - | El cliente solicita por los ganadores del sorteo | 
| `0x06` | Wait | - | El servidor avisa al cliente que espere por el sorteo (todavia no se realizo) | 
| `0x07` | Winners | `1 bytes: length` + `N Documents` | El servidor envia los datos de los ganadores hacia el cliente | 
| - | Document | `1 byte: length` + `Document` | Protocolo de envio de documento dentro del mensaje con `opcode = 0x07` |

- El servidor mantiene una lista con todas las agencias para verificar si ya terminaron de enviar sus apuestas. Una vez que un cliente finaliza el envío de su batch de apuestas, comienza a consultar al servidor por los ganadores.

- La respuesta del servidor puede ser de dos tipos:

    1. `opcode 0x06 (Wait)`: el servidor indica que el sorteo todavía no se realizó. En este caso, el cliente se desconecta temporalmente, permitiendo que el servidor continúe procesando otros mensajes. Pasado un tiempo (en nuestro caso, 3 segundos), el cliente vuelve a conectarse y consulta nuevamente.

    2. `opcode 0x07 (Winners)`: el servidor envía la lista de DNI de los ganadores. El cliente recibe esta información y imprime la cantidad de ganadores obtenidos.

- Correcciones en codigo:
    Se refactorizaron las funciones de protocolo tanto en el cliente como en el servidor, modularizándolas para mejorar la organización del código.

    Ahora el protocolo es un objeto que contiene el socket como atributo, evitando tener que pasarlo como parámetro en cada llamada.

    Además, se elimino codigo repetido y se reorganizaron las funciones, logrando que sean mas legibles y faciles de extender en caso de futuras modificaciones.

### Ej8

Se modifico el cliente, ahora este no se desconecta entre finalizacion de envio de batch y solicitar por los ganadores.

Para modificar el servidor y que pueda procesar mensajes en paralelo, se lanza un thread por cada conexion nueva de un cliente. Este thread ejecuta una funcion `__handle_client_connection`, que procesa los mensajes del cliente.

Aunque Python tiene el GIL (Global Interpreter Lock), que limita la ejecucion simultanea de codigo Python, el multithreading sigue siendo util aqui porque la mayoría de las operaciones del servidor son I/O-bound. Los threads pueden esperar bloqueos de I/O sin bloquear la ejecución de otros threads, permitiendo que el servidor maneje multiples clientes de manera concurrente. Por el contrario, el multithreading suele tener efecto negativo en tareas de alto uso de CPU (CPU-bound), lo que no es nuestro caso.

Para mantener la consistencia de los datos y evitar problemas de concurrencia se utilizan `locks`:
- `_lock_agency_ready`: Protege el acceso a la lista _agency_ready, que indica que agencias han finalizado el envio de sus apuestas.
- `_lock_storage`: Protege el acceso a las funciones de lectura y escritura que persisten las apuestas en el almacenamiento, garantizando que estas operaciones sean seguras ante concurrencia.