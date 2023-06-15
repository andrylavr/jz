import * as THREE from 'three';
import Stats from 'three/addons/libs/stats.module.js';
import {GLTFLoader} from 'three/addons/loaders/GLTFLoader.js';
import {Octree} from 'three/addons/math/Octree.js';
import {OctreeHelper} from 'three/addons/helpers/OctreeHelper.js';
import {Capsule} from 'three/addons/math/Capsule.js';
// import {GUI} from 'three/addons/libs/lil-gui.module.min.js';

const STEPS_PER_FRAME = 5;
const GRAVITY = 30;

class GameAPI {

    start(){
        const clock = new THREE.Clock();
        this.clock = clock;

        const scene = new THREE.Scene();
        this.scene = scene;
        scene.background = new THREE.Color( 0x88ccee );
        // scene.fog = new THREE.Fog( 0x88ccee, 0, 50 );

        const camera = new THREE.PerspectiveCamera( 70, window.innerWidth / window.innerHeight, 0.1, 1000 );
        this.camera = camera;
        camera.rotation.order = 'YXZ';

        const fillLight1 = new THREE.HemisphereLight( 0x8dc1de, 0x00668d, 0.5 );
        fillLight1.position.set( 2, 1, 1 );
        scene.add( fillLight1 );

        const directionalLight = new THREE.DirectionalLight( 0xffffff, 0.8 );
        directionalLight.position.set( - 5, 25, - 1 );
        scene.add( directionalLight );

        const container = document.getElementById( 'container' );

        const renderer = new THREE.WebGLRenderer( { antialias: true } );
        this.renderer = renderer;
        renderer.setPixelRatio( window.devicePixelRatio );
        renderer.setSize( window.innerWidth, window.innerHeight );
        renderer.shadowMap.enabled = true;
        renderer.shadowMap.type = THREE.VSMShadowMap;
        renderer.toneMapping = THREE.ACESFilmicToneMapping;
        container.appendChild( renderer.domElement );

        const stats = new Stats();
        this.stats = stats;
        stats.domElement.style.position = 'absolute';
        stats.domElement.style.top = '0px';
        container.appendChild( stats.domElement );


        const NUM_SPHERES = 100;
        const SPHERE_RADIUS = 0.2;


        const sphereGeometry = new THREE.IcosahedronGeometry( SPHERE_RADIUS, 5 );
        const sphereMaterial = new THREE.MeshLambertMaterial( { color: 0xdede8d } );

        const spheres = [];
        this.spheres = spheres;
        this.sphereIdx = 0;

        for ( let i = 0; i < NUM_SPHERES; i ++ ) {

            const sphere = new THREE.Mesh( sphereGeometry, sphereMaterial );
            // sphere.castShadow = true;
            // sphere.receiveShadow = true;

            scene.add( sphere );

            spheres.push( {
                mesh: sphere,
                collider: new THREE.Sphere( new THREE.Vector3( 0, - 100, 0 ), SPHERE_RADIUS ),
                velocity: new THREE.Vector3()
            } );

        }

        const worldOctree = new Octree();
        this.worldOctree = worldOctree;

        this.playerCollider = new Capsule(new THREE.Vector3(0, 0.35, 0), new THREE.Vector3(0, 1, 0), 0.35);

        this.playerVelocity = new THREE.Vector3();
        this.playerDirection = new THREE.Vector3();

        this.playerOnFloor = false;
        this.mouseTime = 0;

        const keyStates = {};
        this.keyStates = keyStates;

        this.vector1 = new THREE.Vector3();
        this.vector2 = new THREE.Vector3();
        this.vector3 = new THREE.Vector3();

        document.addEventListener( 'keydown', ( event ) => {

            keyStates[ event.code ] = true;

        } );

        document.addEventListener( 'keyup', ( event ) => {

            keyStates[ event.code ] = false;

        } );

        container.addEventListener( 'mousedown', () => {

            document.body.requestPointerLock();

            this.mouseTime = performance.now();

        } );

        document.addEventListener( 'mouseup', () => {

            if ( document.pointerLockElement !== null ) this.throwBall();

        } );

        document.body.addEventListener( 'mousemove', ( event ) => {

            if ( document.pointerLockElement === document.body ) {

                camera.rotation.y -= event.movementX / 500;
                camera.rotation.x -= event.movementY / 500;

            }

        } );

        window.addEventListener( 'resize', () => this.onWindowResize() );



        const loader = new GLTFLoader().setPath( './models/gltf/' );

        loader.load( 'collision-world.glb', ( gltf ) => {

            scene.add( gltf.scene );

            worldOctree.fromGraphNode( gltf.scene );

            gltf.scene.traverse( child => {

                if ( child.isMesh ) {

                    // child.castShadow = true;
                    // child.receiveShadow = true;

                    if ( child.material.map ) {

                        child.material.map.anisotropy = 4;

                    }

                }

            } );

            const helper = new OctreeHelper( worldOctree );
            helper.visible = false;
            scene.add( helper );

            // const gui = new GUI( { width: 200 } );
            // gui.add( { debug: false }, 'debug' ).onChange(value => helper.visible = value)

            this.loaded = true;
        });

    }

    teleportPlayerIfOob() {

        if ( this.camera.position.y <= - 25 ) {

            this.playerCollider.start.set( 0, 0.35, 0 );
            this.playerCollider.end.set( 0, 1, 0 );
            this.playerCollider.radius = 0.35;
            this.camera.position.copy( this.playerCollider.end );
            this.camera.rotation.set( 0, 0, 0 );

        }

    }

    onWindowResize() {

        this.camera.aspect = window.innerWidth / window.innerHeight;
        this.camera.updateProjectionMatrix();

        this.renderer.setSize( window.innerWidth, window.innerHeight );

    }

    throwBall() {

        const sphere = this.spheres[ this.sphereIdx ];

        this.camera.getWorldDirection( this.playerDirection );

        sphere.collider.center.copy( this.playerCollider.end ).addScaledVector( this.playerDirection, this.playerCollider.radius * 1.5 );

        // throw the ball with more force if we hold the button longer, and if we move forward

        const impulse = 15 + 30 * ( 1 - Math.exp( ( this.mouseTime - performance.now() ) * 0.001 ) );

        sphere.velocity.copy( this.playerDirection ).multiplyScalar( impulse );
        sphere.velocity.addScaledVector( this.playerVelocity, 2 );

        this.sphereIdx = ( this.sphereIdx + 1 ) % this.spheres.length;

    }

    playerCollisions() {

        const result = this.worldOctree.capsuleIntersect(this.playerCollider );

        this.playerOnFloor = false;

        if ( result ) {

            this.playerOnFloor = result.normal.y > 0;

            if ( !this.playerOnFloor ) {

                this.playerVelocity.addScaledVector( result.normal, - result.normal.dot( this.playerVelocity ) );

            }

            this.playerCollider.translate( result.normal.multiplyScalar( result.depth ) );

        }

    }

    updatePlayer( deltaTime ) {

        let damping = Math.exp( - 4 * deltaTime ) - 1;

        if ( ! this.playerOnFloor ) {

            this.playerVelocity.y -= GRAVITY * deltaTime;

            // small air resistance
            damping *= 0.1;

        }

        this.playerVelocity.addScaledVector( this.playerVelocity, damping );

        const deltaPosition = this.playerVelocity.clone().multiplyScalar( deltaTime );
        this.playerCollider.translate( deltaPosition );

        this.playerCollisions();

        this.camera.position.copy( this.playerCollider.end );

    }

    playerSphereCollision( sphere ) {

        const center = this.vector1.addVectors(this.playerCollider.start,this.playerCollider.end ).multiplyScalar( 0.5 );

        const sphere_center = sphere.collider.center;

        const r =this.playerCollider.radius + sphere.collider.radius;
        const r2 = r * r;

        // approximation: player = 3 spheres

        for ( const point of [this.playerCollider.start,this.playerCollider.end, center ] ) {

            const d2 = point.distanceToSquared( sphere_center );

            if ( d2 < r2 ) {

                const normal = this.vector1.subVectors( point, sphere_center ).normalize();
                const v1 = this.vector2.copy( normal ).multiplyScalar( normal.dot( this.playerVelocity ) );
                const v2 = this.vector3.copy( normal ).multiplyScalar( normal.dot( sphere.velocity ) );

                this.playerVelocity.add( v2 ).sub( v1 );
                sphere.velocity.add( v1 ).sub( v2 );

                const d = ( r - Math.sqrt( d2 ) ) / 2;
                sphere_center.addScaledVector( normal, - d );

            }

        }

    }

    spheresCollisions() {

        for ( let i = 0, length = this.spheres.length; i < length; i ++ ) {

            const s1 = this.spheres[ i ];

            for ( let j = i + 1; j < length; j ++ ) {

                const s2 = this.spheres[ j ];

                const d2 = s1.collider.center.distanceToSquared( s2.collider.center );
                const r = s1.collider.radius + s2.collider.radius;
                const r2 = r * r;

                if ( d2 < r2 ) {

                    const normal = this.vector1.subVectors( s1.collider.center, s2.collider.center ).normalize();
                    const v1 = this.vector2.copy( normal ).multiplyScalar( normal.dot( s1.velocity ) );
                    const v2 = this.vector3.copy( normal ).multiplyScalar( normal.dot( s2.velocity ) );

                    s1.velocity.add( v2 ).sub( v1 );
                    s2.velocity.add( v1 ).sub( v2 );

                    const d = ( r - Math.sqrt( d2 ) ) / 2;

                    s1.collider.center.addScaledVector( normal, d );
                    s2.collider.center.addScaledVector( normal, - d );

                }

            }

        }

    }

    updateSpheres( deltaTime ) {

        this.spheres.forEach( sphere => {

            sphere.collider.center.addScaledVector( sphere.velocity, deltaTime );

            const result = this.worldOctree.sphereIntersect( sphere.collider );

            if ( result ) {

                sphere.velocity.addScaledVector( result.normal, - result.normal.dot( sphere.velocity ) * 1.5 );
                sphere.collider.center.add( result.normal.multiplyScalar( result.depth ) );

            } else {

                sphere.velocity.y -= GRAVITY * deltaTime;

            }

            const damping = Math.exp( - 1.5 * deltaTime ) - 1;
            sphere.velocity.addScaledVector( sphere.velocity, damping );

            this.playerSphereCollision( sphere );

        } );

        this.spheresCollisions();

        for ( const sphere of this.spheres ) {

            sphere.mesh.position.copy( sphere.collider.center );

        }

    }

    getForwardVector() {

        this.camera.getWorldDirection( this.playerDirection );
        this.playerDirection.y = 0;
        this.playerDirection.normalize();

        return this.playerDirection;

    }

    getSideVector() {

        this.camera.getWorldDirection( this.playerDirection );
        this.playerDirection.y = 0;
        this.playerDirection.normalize();
        this.playerDirection.cross( this.camera.up );

        return this.playerDirection;

    }

    controls( deltaTime ) {

        // gives a bit of air control
        const speedDelta = deltaTime * ( this.playerOnFloor ? 25 : 8 );

        if ( this.keyStates[ 'KeyW' ] ) {

            this.playerVelocity.add( this.getForwardVector().multiplyScalar( speedDelta ) );

        }

        if ( this.keyStates[ 'KeyS' ] ) {

            this.playerVelocity.add( this.getForwardVector().multiplyScalar( - speedDelta ) );

        }

        if ( this.keyStates[ 'KeyA' ] ) {

            this.playerVelocity.add( this.getSideVector().multiplyScalar( - speedDelta ) );

        }

        if ( this.keyStates[ 'KeyD' ] ) {

            this.playerVelocity.add( this.getSideVector().multiplyScalar( speedDelta ) );

        }

        if ( this.playerOnFloor ) {

            if ( this.keyStates[ 'Space' ] ) {

                this.playerVelocity.y = 15;

            }

        }

    }

    animate() {
        if (!this.loaded){
            return;
        }

        const deltaTime = Math.min( 0.05, this.clock.getDelta() ) / STEPS_PER_FRAME;

        // we look for collisions in substeps to mitigate the risk of
        // an object traversing another too quickly for detection.

        for ( let i = 0; i < STEPS_PER_FRAME; i ++ ) {

            this.controls( deltaTime );

            this.updatePlayer( deltaTime );

            this.updateSpheres( deltaTime );

            this.teleportPlayerIfOob();

        }

        this.renderer.render( this.scene, this.camera );

        this.stats.update();

        // requestAnimationFrame(() => this.animate());

    }
}

// window.GameAPIJS = new GameAPI()
new GameAPI()