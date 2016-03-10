<?php
session_start();

require 'twitteroauth/autoload.php';
use Abraham\TwitterOAuth\TwitterOAuth;

$key = '';
$sec = '';

$request_token = [];
$url = '';

if(isset($_SESSION["oauth_token"]))
{
        $connection = new TwitterOAuth($key, $sec, $_SESSION['oauth_token'], $_SESSION['oauth_token_secret']);
}
else
{
        $connection = new TwitterOAuth($key, $sec);
        $request_token = $connection->oauth('oauth/request_token', array('oauth_callback' => 'http://0x.tf/tauth.php'));
        $url = $connection->url('oauth/authorize', array('oauth_token' => $request_token['oauth_token']));

        $_SESSION['oauth_token'] = $request_token['oauth_token'];
        $_SESSION['oauth_token_secret'] = $request_token['oauth_token_secret'];
}
?>

<!DOCTYPE html>
<html>

<head>
        <title>haxalert auth</title>
        <style>html,body{ padding:0; margin:0; font-family: Arial; } pre { background-color: #ccc; padding: 10px; }</style>
</head>

<body>
        <h4>Session</h4>
        <pre><?php var_dump($_SESSION); ?></pre>
        <h4>GET</h4>
        <pre><?php var_dump($_GET); ?></pre>
        <h4>POST</h4>
        <pre><?php var_dump($_POST); ?></pre>
        <h4>Request Token</h4>
        <pre><?php var_dump($request_token); ?></pre>
        <h4>Auth link</h4>
        <a href="<?php echo $url; ?>"><?php echo $url; ?></a>


        <?php
                if (isset($_REQUEST['oauth_token']) && isset($_SESSION['oauth_token'])) {
                        $access_token = $connection->oauth("oauth/access_token", ["oauth_verifier" => $_REQUEST['oauth_verifier']]);
                        echo '<h4>Access Token</h4><pre>';
                        var_dump($access_token);
                        echo '</pre>';

                        $sf = fopen('tokens/' . $access_token['screen_name'] . '.json', 'w');
                        fwrite($sf, json_encode($access_token, JSON_PRETTY_PRINT));
                        flcose($sf);
                        session_destroy();
                }
        ?>
</body>

</html>
