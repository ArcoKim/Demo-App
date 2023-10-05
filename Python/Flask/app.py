from flask import Flask, jsonify, render_template

def create_app():
    app = Flask(__name__)

    @app.route('/healthz')
    def health_check():
        health = {'status': 'ok'}
        return jsonify(health)

    @app.route('/version')
    def version():
        name = 'Kim JeongTae'
        version = '0.0.1'
        return render_template('version.html', name=name, version=version)
    
    return app

if __name__ == '__main__':
    app = create_app()
    app.run(host='0.0.0.0',port=8080)